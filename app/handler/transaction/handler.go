package transaction

import (
	paymentresource "MoneyHook/MoneyHook-API/paymentresource"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
)

type Handler struct {
	transactionStore     transactiondomain.Store
	paymentResourceStore paymentresource.Store
}

func New(
	transactionStore transactiondomain.Store,
	paymentResourceStore paymentresource.Store,
) *Handler {
	return &Handler{
		transactionStore:     transactionStore,
		paymentResourceStore: paymentResourceStore,
	}
}
