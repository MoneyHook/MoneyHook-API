package subcategory

import subcategorydomain "MoneyHook/MoneyHook-API/subcategory"

type Handler struct {
	subCategoryStore subcategorydomain.Store
}

func New(subCategoryStore subcategorydomain.Store) *Handler {
	return &Handler{subCategoryStore: subCategoryStore}
}
