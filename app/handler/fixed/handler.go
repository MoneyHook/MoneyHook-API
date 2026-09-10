package fixed

import (
	fixeddomain "MoneyHook/MoneyHook-API/fixed"
)

type Handler struct {
	fixedStore fixeddomain.Store
}

func New(fixedStore fixeddomain.Store) *Handler {
	return &Handler{fixedStore: fixedStore}
}
