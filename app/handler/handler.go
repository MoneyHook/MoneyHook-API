package handler

import (
	category "MoneyHook/MoneyHook-API/cagegory"
	fixed "MoneyHook/MoneyHook-API/fixed"
	sub_category "MoneyHook/MoneyHook-API/sub_cagegory"
	transaction "MoneyHook/MoneyHook-API/transaction"
)

type Handler struct {
	categoryStore    category.Store
	subCategoryStore sub_category.Store
	transactionStore transaction.Store
	fixedStore       fixed.Store
}

func NewHandler(cs category.Store, scs sub_category.Store, ts transaction.Store, fs fixed.Store) *Handler {
	return &Handler{categoryStore: cs, subCategoryStore: scs, transactionStore: ts, fixedStore: fs}
}
