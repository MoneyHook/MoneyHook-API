package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"gorm.io/gorm"
	"sort"
	"time"
)

// These expressions use the same source as the public projection, including captured payer names.
const householdEntryAmountExpr = "CASE WHEN e.kind='shared' THEN ABS(t.transaction_amount) ELSE CAST((" + householdEntryPayloadExpr + ")->>'amount' AS BIGINT) END"
const householdEntrySignExpr = "CASE WHEN e.kind='shared' THEN CASE WHEN t.transaction_amount < 0 THEN -1 ELSE 1 END ELSE CAST((" + householdEntryPayloadExpr + ")->>'sign' AS INTEGER) END"
const householdEntryExcludedExpr = "CASE WHEN e.kind='shared' THEN false ELSE COALESCE(CAST((" + householdEntryPayloadExpr + ")->>'excluded_from_totals' AS BOOLEAN),false) END"

func householdEntryFactsQuery(tx *gorm.DB, f string) *gorm.DB {
	return householdEntryJoinedQuery(tx).Where("e.household_id = ? AND e.state='active'", f).
		Select("e.entry_id", householdEntryDateExpr+" AS date", householdEntryAmountExpr+" AS amount", householdEntrySignExpr+" AS sign", householdEntryExcludedExpr+" AS excluded")
}

func (s *HouseholdStore) Analytics(ctx context.Context, u, f, month, group string) (*h.Summary, error) {
	tx := s.db.WithContext(ctx)
	if _, _, err := familyAccess(tx, u, f, false); err != nil {
		return nil, err
	}
	if !h.ValidMonth(month) || (group != "overview" && group != "categories" && group != "payers") {
		return nil, h.Invalid
	}
	start, _ := time.Parse("2006-01-02", month)
	idExpr := "CASE WHEN e.kind='shared' THEN CAST(t.category_id AS TEXT) ELSE (" + householdEntryPayloadExpr + ")->>'category_id' END"
	nameExpr := "CASE WHEN e.kind='shared' THEN c.category_name ELSE (" + householdEntryPayloadExpr + ")->>'category_name' END"
	if group == "payers" {
		idExpr = "CASE WHEN e.kind='snapshot' THEN COALESCE((" + householdEntryPayloadExpr + ")->'payer'->>'member_id','common') ELSE COALESCE(CAST(e.payer_member_id AS TEXT),'common') END"
		nameExpr = "CASE WHEN e.kind='snapshot' THEN (" + householdEntryPayloadExpr + ")->'payer'->>'display_name' WHEN e.payer_member_id IS NULL THEN '家族共通' ELSE m.display_name END"
	}
	q := householdEntryFactsQuery(tx, f).
		Select("e.entry_id", householdEntryAmountExpr+" AS amount", householdEntrySignExpr+" AS sign", idExpr+" AS group_id", nameExpr+" AS group_name").
		Where(householdEntryDateExpr+" >= ? AND "+householdEntryDateExpr+" < ?", month, start.AddDate(0, 1, 0).Format("2006-01-02")).Where("NOT (" + householdEntryExcludedExpr + ")")
	// The former descending traversal retained the oldest expense's name for each ID.
	named := tx.Table("(?) AS entries", q).Select("*, FIRST_VALUE(group_name) OVER (PARTITION BY group_id ORDER BY CASE WHEN sign=1 THEN 1 ELSE 0 END, entry_id) AS expense_name")
	var rows []struct {
		GroupID      string
		ExpenseName  string
		Income       int64
		Expense      int64
		ExpenseCount int64
	}
	err := tx.Table("(?) AS named", named).Select(`group_id,expense_name,
 SUM(CASE WHEN sign=1 THEN amount ELSE 0 END) AS income,
 SUM(CASE WHEN sign=1 THEN 0 ELSE amount END) AS expense,
 SUM(CASE WHEN sign=1 THEN 0 ELSE 1 END) AS expense_count`).Group("group_id,expense_name").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := &h.Summary{Groups: []h.SummaryGroup{}}
	for _, row := range rows {
		result.Income += row.Income
		result.Expense += row.Expense
		if row.ExpenseCount > 0 {
			result.Groups = append(result.Groups, h.SummaryGroup{ID: row.GroupID, Name: row.ExpenseName, Amount: row.Expense})
		}
	}
	result.Balance = result.Income - result.Expense
	sort.Slice(result.Groups, func(i, j int) bool { return result.Groups[i].Amount > result.Groups[j].Amount })
	return result, nil
}

func (s *HouseholdStore) Duplicates(ctx context.Context, u, f string, filter h.DuplicateFilter) ([]h.Entry, error) {
	tx := s.db.WithContext(ctx)
	_, viewer, err := familyAccess(tx, u, f, false)
	if err != nil {
		return nil, err
	}
	date, err := time.Parse("2006-01-02", filter.Date)
	if err != nil || date.Format("2006-01-02") != filter.Date || filter.Amount < 1 || filter.Amount > 9999999 || (filter.Sign != -1 && filter.Sign != 1) || (filter.Exclude != "" && !h.ID(filter.Exclude)) {
		return nil, h.Invalid
	}
	q := householdEntryFactsQuery(tx, f)
	if filter.Payer == "common" {
		q = q.Where("e.payer_member_id IS NULL")
	} else if filter.Payer != "" {
		if !h.ID(filter.Payer) {
			return nil, h.Invalid
		}
		q = q.Where("e.payer_member_id = ?", filter.Payer)
	}
	matches := tx.Table("(?) AS candidates", q).Select("entry_id").Where("date = ? AND amount = ? AND sign = ? AND NOT excluded", filter.Date, filter.Amount, filter.Sign)
	if filter.Exclude != "" {
		matches = matches.Where("entry_id <> ?", filter.Exclude)
	}
	var rows []householdEntryReadRecord
	if err := householdEntryReadQuery(tx).Where("e.household_id = ? AND e.entry_id IN (?)", f, matches).Order("e.entry_id DESC").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := []h.Entry{}
	for i := range rows {
		entry, err := projectHouseholdEntry(&rows[i], viewer)
		if err != nil {
			return nil, err
		}
		out = append(out, *entry)
	}
	return out, nil
}
