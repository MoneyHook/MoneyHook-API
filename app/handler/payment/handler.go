package payment

import paymentresource "MoneyHook/MoneyHook-API/paymentresource"

type Handler struct {
	paymentResourceStore paymentresource.Store
}

func New(paymentResourceStore paymentresource.Store) *Handler {
	return &Handler{paymentResourceStore: paymentResourceStore}
}
