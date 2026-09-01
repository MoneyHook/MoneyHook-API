package budget

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetBudget(userNo string, month string) (*model.Budget, error)
	UpsertBudget(*model.Budget) error
}
