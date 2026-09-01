package transaction

import (
	paymentresource "MoneyHook/MoneyHook-API/paymentresource"
	subcategorydomain "MoneyHook/MoneyHook-API/subcategory"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
)

type Handler struct {
	transactionStore     transactiondomain.Store
	subCategoryStore     subcategorydomain.Store
	paymentResourceStore paymentresource.Store
}

func New(
	transactionStore transactiondomain.Store,
	subCategoryStore subcategorydomain.Store,
	paymentResourceStore paymentresource.Store,
) *Handler {
	return &Handler{
		transactionStore:     transactionStore,
		subCategoryStore:     subCategoryStore,
		paymentResourceStore: paymentResourceStore,
	}
}
