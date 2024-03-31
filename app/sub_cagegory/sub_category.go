package sub_category

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetSubCategoryList(userId int, categoryId int) *[]model.SubCategory
	CreateSubCategory(*model.SubCategoryModel) *model.SubCategoryModel
	HideSubCategory(*model.EditSubCategoryModel)
	ExposeSubCategory(*model.EditSubCategoryModel)
}
