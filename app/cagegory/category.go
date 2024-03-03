package category

import "example.com/m/model"

type Store interface {
	GetCategoryList() *model.CategoryList
	GetCategoryWithSubCategoryList() *model.CategoryWithSubCategoryList
}
