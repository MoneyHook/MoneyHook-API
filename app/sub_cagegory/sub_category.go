package sub_category

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetSubCategoryList(userId string, categoryId string) *[]model.SubCategory
	CreateSubCategory(*model.SubCategoryModel) error
	HideSubCategory(*model.EditSubCategoryModel) error
	ExposeSubCategory(*model.EditSubCategoryModel) error
	FindByName(*model.SubCategoryModel) bool
}
