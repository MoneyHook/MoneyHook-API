package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/model"
	"time"
)

func (ts *TransactionStore) ListV1Transactions(user, month, sharing, cursor string) ([]model.V1Transaction, *string, error) {
	if !h.ValidMonth(month) || (cursor != "" && !h.ID(cursor)) {
		return nil, nil, h.Invalid
	}
	start, _ := time.Parse("2006-01-02", month)
	q := ts.db.Table("transaction t").Select("t.transaction_id").Where("t.user_no = ? AND t.deleted_at IS NULL AND t.transaction_date >= ? AND t.transaction_date < ?", user, month, start.AddDate(0, 1, 0).Format("2006-01-02"))
	shared := "EXISTS (SELECT 1 FROM household_entry he WHERE he.source_transaction_id=t.transaction_id AND he.kind='shared' AND he.state='active')"
	switch sharing {
	case "shared":
		q = q.Where(shared)
	case "private":
		q = q.Where("NOT " + shared)
	case "", "all":
	default:
		return nil, nil, h.Invalid
	}
	if cursor != "" {
		q = q.Where("t.transaction_id < ?", cursor)
	}
	var ids []uint64
	if e := q.Order("t.transaction_id DESC").Limit(101).Scan(&ids).Error; e != nil {
		return nil, nil, e
	}
	var next *string
	if len(ids) > 100 {
		v := stringID(ids[99])
		next = &v
		ids = ids[:100]
	}
	out := []model.V1Transaction{}
	for _, id := range ids {
		t, e := ts.GetV1Transaction(user, stringID(id))
		if e != nil {
			return nil, nil, e
		}
		out = append(out, *t)
	}
	return out, next, nil
}
