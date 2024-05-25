package handler

import (
	category "MoneyHook/MoneyHook-API/cagegory"
	fixed "MoneyHook/MoneyHook-API/fixed"
	sub_category "MoneyHook/MoneyHook-API/sub_cagegory"
	transaction "MoneyHook/MoneyHook-API/transaction"
	user "MoneyHook/MoneyHook-API/user"

	"firebase.google.com/go/auth"
)

type Handler struct {
	firebaseClient   *auth.Client
	userStore        user.Store
	transactionStore transaction.Store
	fixedStore       fixed.Store
	categoryStore    category.Store
	subCategoryStore sub_category.Store
}

func NewHandler(
	fc *auth.Client,
	us user.Store,
	ts transaction.Store,
	fs fixed.Store,
	cs category.Store,
	scs sub_category.Store,
) *Handler {
	return &Handler{
		firebaseClient:   fc,
		categoryStore:    cs,
		subCategoryStore: scs,
		transactionStore: ts,
		fixedStore:       fs,
		userStore:        us,
	}
}
