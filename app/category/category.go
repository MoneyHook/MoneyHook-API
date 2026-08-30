package category

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetCategoryList() *[]model.Category
	GetCategoryWithSubCategoryList(userId string) *[]model.CategoryWithSubCategory
}
