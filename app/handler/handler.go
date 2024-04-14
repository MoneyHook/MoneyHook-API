package handler

import (
	category "MoneyHook/MoneyHook-API/cagegory"
	fixed "MoneyHook/MoneyHook-API/fixed"
	sub_category "MoneyHook/MoneyHook-API/sub_cagegory"
	transaction "MoneyHook/MoneyHook-API/transaction"
	user "MoneyHook/MoneyHook-API/user"
)

type Handler struct {
	userStore        user.Store
	transactionStore transaction.Store
	fixedStore       fixed.Store
	categoryStore    category.Store
	subCategoryStore sub_category.Store
}

func NewHandler(
	us user.Store,
	ts transaction.Store,
	fs fixed.Store,
	cs category.Store,
	scs sub_category.Store,
) *Handler {
	return &Handler{
		categoryStore:    cs,
		subCategoryStore: scs,
		transactionStore: ts,
		fixedStore:       fs,
		userStore:        us,
	}
}
