package subcategory

import (
	"MoneyHook/MoneyHook-API/model"
	"errors"
)

var (
	ErrResolveFailed = errors.New("subcategory resolution failed")
	ErrNotFound      = errors.New("subcategory not found")
)

type Store interface {
	GetSubCategoryList(userId string, categoryId string) *[]model.SubCategory
	HideSubCategory(*model.EditSubCategoryModel) error
	ExposeSubCategory(*model.EditSubCategoryModel) error
}
