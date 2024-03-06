package store

import (
	"MoneyHook/MoneyHook-API/model"
	"fmt"

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
	cs.db.Table("sub_category sc").Joins("LEFT JOIN hidden_sub_category hsc ON sc.sub_category_id = hsc.sub_category_id	").Where("sc.user_no IN ?", []int{1, userId}).Where("sc.category_id = ?", categoryId).Where("hsc.sub_category_id is NULL").Find(&sub_category_list)

	for _, v := range sub_category_list {
		fmt.Println(v)
	}

	return &sub_category_list
}
