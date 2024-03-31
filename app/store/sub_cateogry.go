package store

import (
	"MoneyHook/MoneyHook-API/model"

	"gorm.io/gorm"
)

type SubCategoryStore struct {
	db *gorm.DB
}

func NewSubCategoryStore(db *gorm.DB) *SubCategoryStore {
	return &SubCategoryStore{db: db}
}

func (cs *SubCategoryStore) GetSubCategoryList(userId int, categoryId int) *[]model.SubCategory {
	var sub_category_list []model.SubCategory
	cs.db.Table("sub_category sc").
		Joins("LEFT JOIN hidden_sub_category hsc ON sc.sub_category_id = hsc.sub_category_id").
		Where("sc.user_no IN ?", []int{1, userId}).
		Where("sc.category_id = ?", categoryId).
		Where("hsc.sub_category_id is NULL").
		Find(&sub_category_list)

	return &sub_category_list
}

func (cs *SubCategoryStore) CreateSubCategory(subCategory *model.SubCategoryModel) *model.SubCategoryModel {
	cs.db.Table("sub_category").Create(&subCategory)
	return subCategory
}

func (cs *SubCategoryStore) HideSubCategory(subCategory *model.EditSubCategoryModel) {
	cs.db.Table("hidden_sub_category").Create(map[string]interface{}{
		"user_no":         subCategory.UserId,
		"sub_category_id": subCategory.SubCategoryId,
	})
}

func (cs *SubCategoryStore) ExposeSubCategory(subCategory *model.EditSubCategoryModel) {
	cs.db.Table("hidden_sub_category").
		Where("user_no = ?", subCategory.UserId).
		Where("sub_category_id = ?", subCategory.SubCategoryId).
		Delete(&subCategory)
}
