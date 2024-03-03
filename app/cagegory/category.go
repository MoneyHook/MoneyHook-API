package category

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetCategoryList() *model.CategoryList
	GetCategoryWithSubCategoryList() *model.CategoryWithSubCategoryList
}
