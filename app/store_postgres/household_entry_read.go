package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"time"
)

type householdEntryRecord struct {
	ID                  uint64 `gorm:"column:entry_id;primaryKey;autoIncrement"`
	HouseholdID         uint64
	Kind                string
	SourceTransactionID *uint64
	PayerMemberID       *uint64
	PaymentID           *uint64
	SubCategoryID       *uint64
	State               string
	Version             int64
	CreatedBy           uint64
	UpdatedBy           uint64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (householdEntryRecord) TableName() string { return "household_entry" }
func entryRecord(tx *gorm.DB, f, id string) (*householdEntryRecord, error) {
	var r householdEntryRecord
	e := tx.Where("household_id = ? AND entry_id = ?", f, id).Take(&r).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return nil, h.NotFound
	}
	return &r, e
}
func entryProjection(tx *gorm.DB, r *householdEntryRecord, viewer *h.Member) (*h.Entry, error) {
	var out h.Entry
	if r.Kind == "shared" {
		var v struct {
			Name         string
			Amount       int64
			Date         string
			Time         *string
			CategoryID   string
			CategoryName string
			Fixed        bool
			Owner        string
			Version      int64
		}
		e := tx.Table("transaction t").Select("t.transaction_name AS name,t.transaction_amount AS amount,TO_CHAR(t.transaction_date,'YYYY-MM-DD') AS date,LEFT(CAST(t.transaction_time AS TEXT),5) AS time,CAST(t.category_id AS TEXT) AS category_id,c.category_name,t.fixed_flg AS fixed,CAST(t.user_no AS TEXT) AS owner,t.version").Joins("JOIN category c ON c.category_id=t.category_id").Where("t.transaction_id = ? AND t.deleted_at IS NULL", r.SourceTransactionID).Take(&v).Error
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, h.NotFound
		}
		if e != nil {
			return nil, e
		}
		out.TransactionInput = h.TransactionInput{Name: v.Name, Date: v.Date, Time: v.Time, CategoryID: v.CategoryID, Fixed: v.Fixed, Amount: v.Amount, Sign: 1}
		if v.Amount < 0 {
			out.Amount = -v.Amount
			out.Sign = -1
		}
		out.SignedAmount = v.Amount
		out.CategoryName = v.CategoryName
		if viewer != nil && stringID(viewer.UserNo) == v.Owner {
			id := stringID(*r.SourceTransactionID)
			out.SourceTransactionID = &id
			out.SourceVersion = &v.Version
			out.Permissions = h.Permissions{Edit: viewer.State == "active", Delete: viewer.State == "active", Unshare: viewer.State == "active"}
		}
	} else {
		var d struct {
			Payload          json.RawMessage
			CorrectedPayload json.RawMessage
			CapturedAt       *time.Time
		}
		if e := tx.Table("household_entry_data").Where("entry_id = ?", r.ID).Take(&d).Error; e != nil {
			return nil, e
		}
		if e := json.Unmarshal(d.Payload, &out); e != nil {
			return nil, e
		}
		out.CapturedAt = d.CapturedAt
		if r.Kind == "snapshot" && len(d.CorrectedPayload) > 0 {
			if e := json.Unmarshal(d.CorrectedPayload, &out); e != nil {
				return nil, e
			}
			out.CapturedAt = d.CapturedAt
			out.Corrected = true
		}
		out.Permissions = h.Permissions{Edit: r.Kind == "proxy" && viewer != nil && viewer.State == "active", Delete: r.Kind == "proxy" && viewer != nil && viewer.State == "active", Correct: r.Kind == "snapshot" && viewer != nil && viewer.State == "active"}
	}
	out.ID = stringID(r.ID)
	out.Kind = r.Kind
	out.Version = r.Version
	out.CreatedBy = stringID(r.CreatedBy)
	out.UpdatedBy = stringID(r.UpdatedBy)
	out.UpdatedAt = r.UpdatedAt
	if r.Kind != "snapshot" {
		if e := projectEntryReferences(tx, r, &out); e != nil {
			return nil, e
		}
	}
	// Never expose personal references in a family response.
	out.PaymentID = nil
	out.SubCategoryID = ""
	out.SubCategoryName = ""
	return &out, nil
}
func publicEntry(e *h.Entry) json.RawMessage {
	v := *e
	v.SourceTransactionID = nil
	v.SourceVersion = nil
	v.Permissions = h.Permissions{}
	b, _ := json.Marshal(v)
	return b
}
func (s *HouseholdStore) Entry(ctx context.Context, u, f, id string) (*h.Entry, error) {
	tx := s.db.WithContext(ctx)
	_, m, e := familyAccess(tx, u, f, false)
	if e != nil {
		return nil, e
	}
	r, e := entryRecord(tx, f, id)
	if e != nil {
		return nil, e
	}
	if r.State != "active" {
		return nil, h.NotFound
	}
	return entryProjection(tx, r, m)
}
func (s *HouseholdStore) Entries(ctx context.Context, u, f string, filter h.Filter) (*h.EntryPage, error) {
	tx := s.db.WithContext(ctx)
	_, m, e := familyAccess(tx, u, f, false)
	if e != nil {
		return nil, e
	}
	if !h.ValidMonth(filter.Month) {
		return nil, h.Invalid
	}
	date, _ := time.Parse("2006-01-02", filter.Month)
	end := date.AddDate(0, 1, 0).Format("2006-01-02")
	q := tx.Table("household_entry e").Select("e.*").Joins(`LEFT JOIN "transaction" t ON e.kind='shared' AND t.transaction_id=e.source_transaction_id AND t.deleted_at IS NULL`).Joins("LEFT JOIN household_entry_data d ON d.entry_id=e.entry_id").Where("e.household_id = ? AND e.state='active'", f)
	dateExpr := `CASE WHEN e.kind='shared' THEN TO_CHAR(t.transaction_date,'YYYY-MM-DD') WHEN e.kind='snapshot' THEN COALESCE(d.corrected_payload->>'transaction_date',d.payload->>'transaction_date') ELSE d.payload->>'transaction_date' END`
	q = q.Where(dateExpr+" >= ? AND "+dateExpr+" < ?", filter.Month, end)
	if filter.Kind != "" {
		if filter.Kind != "shared" && filter.Kind != "proxy" && filter.Kind != "snapshot" {
			return nil, h.Invalid
		}
		q = q.Where("e.kind = ?", filter.Kind)
	}
	if filter.Payer == "common" {
		q = q.Where("e.payer_member_id IS NULL")
	} else if filter.Payer != "" {
		if !h.ID(filter.Payer) {
			return nil, h.Invalid
		}
		q = q.Where("e.payer_member_id = ?", filter.Payer)
	}
	if filter.Cursor != "" {
		if !h.ID(filter.Cursor) {
			return nil, h.Invalid
		}
		q = q.Where("e.entry_id < ?", filter.Cursor)
	}
	var rows []householdEntryRecord
	if e = q.Order("e.entry_id DESC").Limit(101).Find(&rows).Error; e != nil {
		return nil, e
	}
	out := &h.EntryPage{Entries: []h.Entry{}}
	if len(rows) > 100 {
		cursor := stringID(rows[99].ID)
		out.NextCursor = &cursor
		rows = rows[:100]
	}
	for i := range rows {
		entry, e := entryProjection(tx, &rows[i], m)
		if e != nil {
			return nil, e
		}
		out.Entries = append(out.Entries, *entry)
	}
	return out, nil
}
func projectEntryReferences(tx *gorm.DB, r *householdEntryRecord, out *h.Entry) error {
	out.Payer = h.Payer{Kind: "common", DisplayName: "家族共通"}
	if r.PayerMemberID != nil {
		var m h.Member
		if e := tx.First(&m, "member_id = ?", *r.PayerMemberID).Error; e != nil {
			return e
		}
		id := stringID(m.ID)
		out.Payer = h.Payer{Kind: "member", MemberID: &id, DisplayName: m.DisplayName, State: m.State}
	}
	if r.PaymentID != nil {
		var n string
		if e := tx.Table("household_payment").Select("name").Where("payment_id = ?", *r.PaymentID).Scan(&n).Error; e != nil {
			return e
		}
		id := stringID(*r.PaymentID)
		out.HouseholdPaymentID = &id
		out.HouseholdPaymentName = &n
	} else {
		out.HouseholdPaymentID = nil
		out.HouseholdPaymentName = nil
	}
	if r.SubCategoryID != nil {
		var n string
		if e := tx.Table("household_sub_category").Select("name").Where("sub_category_id = ?", *r.SubCategoryID).Scan(&n).Error; e != nil {
			return e
		}
		id := stringID(*r.SubCategoryID)
		out.HouseholdSubCategoryID = &id
		out.HouseholdSubCategoryName = &n
	} else {
		out.HouseholdSubCategoryID = nil
		out.HouseholdSubCategoryName = nil
	}

	return nil
}
