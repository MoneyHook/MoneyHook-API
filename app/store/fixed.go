package store

import (
	"MoneyHook/MoneyHook-API/model"

	"gorm.io/gorm"
)

type FixedStore struct {
	db *gorm.DB
}

func NewFixedStore(db *gorm.DB) *FixedStore {
	return &FixedStore{db: db}
}

func (fs *FixedStore) GetFixedData(userId int) *[]model.GetFixed {
	var fixed_list []model.GetFixed

	fs.db.Unscoped().
		Select("mt.monthly_transaction_id",
			"mt.monthly_transaction_name",
			"ABS(mt.monthly_transaction_amount) AS monthly_transaction_amount",
			"(CASE WHEN mt.monthly_transaction_amount > 0 THEN 1 ELSE -1 END) AS monthly_transaction_sign",
			"mt.monthly_transaction_date",
			"mt.category_id",
			"c.category_name",
			"mt.sub_category_id",
			"sc.sub_category_name").
		Table("monthly_transaction mt").
		Joins("INNER JOIN category c ON c.category_id = mt.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = mt.sub_category_id").
		Where("mt.user_no = ?", userId).
		Where("include_flg = TRUE").
		Find(&fixed_list)

	return &fixed_list
}

func (fs *FixedStore) GetFixedDeletedData(userId int) *[]model.GetDeletedFixed {
	var fixed_list []model.GetDeletedFixed

	fs.db.Unscoped().
		Select("mt.monthly_transaction_id",
			"mt.monthly_transaction_name",
			"ABS(mt.monthly_transaction_amount) AS monthly_transaction_amount",
			"mt.monthly_transaction_date",
			"c.category_name",
			"sc.sub_category_name").
		Table("monthly_transaction mt").
		Joins("INNER JOIN category c ON c.category_id = mt.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = mt.sub_category_id").
		Where("mt.user_no = ?", userId).
		Where("include_flg = FALSE").
		Find(&fixed_list)

	return &fixed_list
}
