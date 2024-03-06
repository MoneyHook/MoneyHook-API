package handler

import (
	category "MoneyHook/MoneyHook-API/cagegory"
	sub_category "MoneyHook/MoneyHook-API/sub_cagegory"
)

type Handler struct {
	categoryStore    category.Store
	subCategoryStore sub_category.Store
}

func NewHandler(cs category.Store, scs sub_category.Store) *Handler {
	return &Handler{categoryStore: cs, subCategoryStore: scs}
}
