package fixed

import (
	fixeddomain "MoneyHook/MoneyHook-API/fixed"
	subcategorydomain "MoneyHook/MoneyHook-API/subcategory"
)

type Handler struct {
	fixedStore       fixeddomain.Store
	subCategoryStore subcategorydomain.Store
}

func New(fixedStore fixeddomain.Store, subCategoryStore subcategorydomain.Store) *Handler {
	return &Handler{fixedStore: fixedStore, subCategoryStore: subCategoryStore}
}
