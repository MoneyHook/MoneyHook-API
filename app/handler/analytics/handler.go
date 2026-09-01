package analytics

import transactiondomain "MoneyHook/MoneyHook-API/transaction"

type Handler struct {
	transactionStore transactiondomain.Store
}

func New(transactionStore transactiondomain.Store) *Handler {
	return &Handler{transactionStore: transactionStore}
}
