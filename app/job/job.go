package job

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	SelectMonthlyTransaction(date int, isLastDay bool) (*[]model.JobMonthlyTransaction, error)
	InsertTransaction(*[]model.JobTransaction) error
}
