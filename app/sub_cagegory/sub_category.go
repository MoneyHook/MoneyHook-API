package sub_category

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetSubCategoryList(userId int, categoryId int) *[]model.SubCategory
	CreateSubCategory(*model.SubCategoryModel) error
	HideSubCategory(*model.EditSubCategoryModel) error
	ExposeSubCategory(*model.EditSubCategoryModel) error
	FindByName(*model.SubCategoryModel) bool
}
