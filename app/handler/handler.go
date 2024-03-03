package handler

import category "MoneyHook/MoneyHook-API/cagegory"

type Handler struct {
	categoryStore category.Store
}

func NewHandler(cs category.Store) *Handler {
	return &Handler{categoryStore: cs}
}
