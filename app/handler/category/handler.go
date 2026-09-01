package category

import categorydomain "MoneyHook/MoneyHook-API/category"

type Handler struct {
	categoryStore categorydomain.Store
}

func New(categoryStore categorydomain.Store) *Handler {
	return &Handler{categoryStore: categoryStore}
}
