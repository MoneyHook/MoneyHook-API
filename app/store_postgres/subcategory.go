package store_postgres

import (
	"MoneyHook/MoneyHook-API/model"
	subcategorydomain "MoneyHook/MoneyHook-API/subcategory"
	"errors"
	"fmt"
	"gorm.io/gorm/clause"
	"strconv"

	"gorm.io/gorm"
)

type SubCategoryStore struct {
	db *gorm.DB
}

func NewSubCategoryStore(db *gorm.DB) *SubCategoryStore {
	return &SubCategoryStore{db: db}
}

func (cs *SubCategoryStore) GetSubCategoryList(userId string, categoryId string) *[]model.SubCategory {
	var sub_category_list []model.SubCategory
	cs.db.Table("sub_category sc").
		Joins("LEFT JOIN hidden_sub_category hsc ON sc.sub_category_id = hsc.sub_category_id AND hsc.user_no = ?", userId).
		Where("sc.user_no IN ?", []string{"1", userId}).
		Where("sc.category_id = ?", categoryId).
		Where("hsc.sub_category_id is NULL").
		Find(&sub_category_list)

	return &sub_category_list
}

func (cs *SubCategoryStore) HideSubCategory(subCategory *model.EditSubCategoryModel) error {
	if err := cs.requireAccessibleSubCategory(subCategory); err != nil {
		return err
	}
	return cs.db.Table("hidden_sub_category").Clauses(clause.OnConflict{DoNothing: true}).Create(map[string]interface{}{
		"user_no":         subCategory.UserId,
		"sub_category_id": subCategory.SubCategoryId,
	}).Error
}

func (cs *SubCategoryStore) ExposeSubCategory(subCategory *model.EditSubCategoryModel) error {
	if err := cs.requireAccessibleSubCategory(subCategory); err != nil {
		return err
	}
	return cs.db.Table("hidden_sub_category").
		Where("user_no = ?", subCategory.UserId).
		Where("sub_category_id = ?", subCategory.SubCategoryId).
		Delete(&subCategory).Error
}

func (cs *SubCategoryStore) requireAccessibleSubCategory(subCategory *model.EditSubCategoryModel) error {
	var count int64
	if err := cs.db.Table("sub_category").
		Where("sub_category_id = ?", subCategory.SubCategoryId).
		Where("user_no IN ?", []string{"1", subCategory.UserId}).
		Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return subcategorydomain.ErrNotFound
	}
	return nil
}

// resolveWriteSubCategory must run inside the caller's write transaction.
func resolveWriteSubCategory(tx *gorm.DB, userID, categoryID, subCategoryID, name string) (string, error) {
	if subCategoryID != "" {
		return subCategoryID, nil
	}
	subCategory := model.SubCategoryModel{UserNo: userID, CategoryId: categoryID, SubCategoryName: name}
	query := func() *gorm.DB {
		return tx.Table("sub_category").Where("user_no = ? AND category_id = ? AND sub_category_name = ?", userID, categoryID, name)
	}
	err := query().Take(&subCategory).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = tx.Table("sub_category").Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_no"}, {Name: "category_id"}, {Name: "sub_category_name"}}, DoNothing: true,
		}).Create(&subCategory).Error
		if err == nil {
			err = query().Take(&subCategory).Error
		}
	}
	if err != nil {
		return "", fmt.Errorf("%w: %w", subcategorydomain.ErrResolveFailed, err)
	}
	return strconv.FormatInt(subCategory.SubCategoryId, 10), nil
}
