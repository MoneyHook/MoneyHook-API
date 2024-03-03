package store

import (
	"fmt"

	"example.com/m/model"
	"gorm.io/gorm"
)

type CategoryStore struct {
	db *gorm.DB
}

func NewCategoryStore(db *gorm.DB) *CategoryStore {
	return &CategoryStore{db: db}
}

func (cs *CategoryStore) GetCategoryList() *model.CategoryList {
	var category_list model.CategoryList

	// category_id := 100
	// cs.db.Debug().Select("category_id, category_name").Where("category_id = ?", category_id).Find(&category_list.CategoryList)
	cs.db.Model(&model.Category{})
	// cs.db.Select("category_id, category_name")
	cs.db.Table("category").Debug().Find(&category_list.CategoryList)
	// cs.db.Debug().Find(&category_list.CategoryList)

	// if err != nil {
	// 	fmt.Println("******")
	// 	fmt.Println("Passed")
	// 	fmt.Println("******")
	// 	// if errors.Is(err, gorm.ErrRecordNotFound) {
	// 	// 	// データが存在しない場合の処理
	// 	// } else {
	// 	// 	// その他のエラー処理
	// 	// }
	// }
	fmt.Println(category_list)
	return &category_list
}

func (cs *CategoryStore) GetCategoryWithSubCategoryList() *model.CategoryWithSubCategoryList {
	var result_list model.CategoryWithSubCategoryList

	// cs.db.Preload("SubCategoryList").Find(&result_list.CategoryWithSubCategoryList)
	// cs.db.Select("category.category_id, category.category_name, sub_category.sub_category_id, sub_category.sub_category_name, hidden_sub_category.sub_category_id IS NULL as enable")
	// cs.db.Model(&model.CategoryWithSubCategory{})
	// cs.db.Joins("JOIN sub_category ON category.category_id = sub_category.category_id")
	// cs.db.Joins("LEFT JOIN hidden_sub_category ON sub_category.sub_category_id = hidden_sub_category.sub_category_id")
	// cs.db.Select("category.category_id, category.category_name, sub_category.sub_category_id, sub_category.sub_category_name")

	// cs.db.Unscoped()
	// cs.db.Model(&model.CategoryWithSubCategory{})
	// cs.db.Preload("SubCategoryList")
	// cs.db.Debug().Find(&result_list.CategoryWithSubCategoryList)

	// cs.db.Unscoped().Model(&model.CategoryWithSubCategory{}).Preload("SubCategoryList").Debug().Find(&result_list.CategoryWithSubCategoryList)
	cs.db.Table("category").Find(&result_list.CategoryWithSubCategoryList)

	for i, v := range result_list.CategoryWithSubCategoryList {

		cs.db.Table("sub_category").Where("category_id = ?", v.CategoryId).Where("user_no IN ? ", []int{1, 2}).Find(&result_list.CategoryWithSubCategoryList[i].SubCategoryList)

		// 	cs.db.Table("sub_category")
		// 	cs.db.Where("category_id = ?", v.CategoryId)
		// 	cs.db.Where("user_no IN ? ", []int{1, 2})
		// 	cs.db.Find(&v.SubCategoryList)

	}

	// for _, c := range result_list.CategoryWithSubCategoryList {
	// 	fmt.Print("category: ", c.CategoryId, c.CategoryName)
	// 	for _, s := range c.SubCategoryList {
	// 		fmt.Print("sub_category: ", s.SubCategoryId, s.SubCategoryName)
	// 	}
	// 	fmt.Println()
	// }

	return &result_list
}
