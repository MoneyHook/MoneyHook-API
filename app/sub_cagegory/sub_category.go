package sub_category

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetSubCategoryList(userId int, categoryId int) *[]model.SubCategory
	CreateSubCategory(*model.SubCategoryModel) error
	HideSubCategory(*model.EditSubCategoryModel)
	ExposeSubCategory(*model.EditSubCategoryModel)
	FindByName(*model.SubCategoryModel) bool
}
