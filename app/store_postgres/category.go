package store_postgres

import (
	"MoneyHook/MoneyHook-API/model"

	"gorm.io/gorm"
)

type CategoryStore struct {
	db *gorm.DB
}

func NewCategoryStore(db *gorm.DB) *CategoryStore {
	return &CategoryStore{db: db}
}

func (cs *CategoryStore) GetCategoryList() *[]model.Category {
	var category_list []model.Category

	cs.db.Model(&model.Category{})
	cs.db.Table("category").
		Order("order_num").
		Find(&category_list)

	return &category_list
}

func (cs *CategoryStore) GetCategoryWithSubCategoryList(userId string) *[]model.CategoryWithSubCategory {
	var result []model.CategoryWithSubCategory

	cs.db.Table("category").Select("category_id", "category_name").Order("order_num").Find(&result)

	if len(result) == 0 {
		return &result
	}
	ids := make([]string, len(result))
	positions := make(map[string]int, len(result))
	for i := range result {
		ids[i] = result[i].CategoryId
		positions[ids[i]] = i
		result[i].SubCategoryList = []model.SubCategoryWithEnable{}
	}
	var rows []struct {
		CategoryId                  string
		model.SubCategoryWithEnable `gorm:"embedded"`
	}
	cs.db.Table("sub_category sc").
		Select("sc.category_id", "sc.sub_category_id", "sc.sub_category_name", "hsc.sub_category_id IS NULL as enable").
		Joins("LEFT JOIN hidden_sub_category hsc ON sc.sub_category_id = hsc.sub_category_id AND hsc.user_no = ?", userId).
		Where("sc.category_id IN ?", ids).
		Where("sc.user_no IN ?", []string{"1", userId}).
		Find(&rows)
	for _, row := range rows {
		i := positions[row.CategoryId]
		result[i].SubCategoryList = append(result[i].SubCategoryList, row.SubCategoryWithEnable)
	}

	return &result
}
