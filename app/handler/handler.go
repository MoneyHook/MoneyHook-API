package handler

import category "example.com/m/cagegory"

type Handler struct {
	categoryStore category.Store
}

func NewHandler(cs category.Store) *Handler {
	return &Handler{categoryStore: cs}
}
