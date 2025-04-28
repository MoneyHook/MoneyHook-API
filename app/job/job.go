package job

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	SelectMonthlyTransaction(date int, is_last_day bool) *[]model.JobMonthlyTransaction
	InsertTransaction(*[]model.JobTransaction) error
}
