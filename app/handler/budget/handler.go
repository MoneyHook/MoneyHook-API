package budget

import budgetdomain "MoneyHook/MoneyHook-API/budget"

type Handler struct {
	budgetStore budgetdomain.Store
}

func New(budgetStore budgetdomain.Store) *Handler {
	return &Handler{budgetStore: budgetStore}
}
