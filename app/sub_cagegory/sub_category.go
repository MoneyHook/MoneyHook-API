package sub_category

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetSubCategoryList(userId int, categoryId int) *[]model.SubCategory
}
