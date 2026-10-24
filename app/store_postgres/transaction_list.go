package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/model"
	"time"
)

const v1TransactionPageSize = 500

func (ts *TransactionStore) ListV1Transactions(user, month, sharing, cursor string) ([]model.V1Transaction, *string, error) {
	if !h.ValidMonth(month) || (cursor != "" && !h.ID(cursor)) {
		return nil, nil, h.Invalid
	}
	start, _ := time.Parse("2006-01-02", month)
	q := postgresV1TransactionQuery(ts.db, user).
		Where("t.transaction_date >= ? AND t.transaction_date < ?", month, start.AddDate(0, 1, 0).Format("2006-01-02"))
	switch sharing {
	case "shared":
		q = q.Where(v1TransactionSharedExists)
	case "private":
		q = q.Where("NOT " + v1TransactionSharedExists)
	case "", "all":
	default:
		return nil, nil, h.Invalid
	}
	if cursor != "" {
		q = q.Where("t.transaction_id < ?", cursor)
	}
	out := []model.V1Transaction{}
	if e := q.Order("t.transaction_id DESC").Limit(v1TransactionPageSize + 1).Scan(&out).Error; e != nil {
		return nil, nil, e
	}
	var next *string
	if len(out) > v1TransactionPageSize {
		v := out[v1TransactionPageSize-1].TransactionId
		next = &v
		out = out[:v1TransactionPageSize]
	}
	return out, next, nil
}
