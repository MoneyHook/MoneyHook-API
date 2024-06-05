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

func (ts *FixedStore) AddFixed(monthlyTransaction *model.AddFixed) error {
	paymentId := interface{}(monthlyTransaction.PaymentId)
	if monthlyTransaction.PaymentId == 0 {
		paymentId = nil
	}

	return ts.db.Table("monthly_transaction").Create(map[string]interface{}{
		"user_no":                    monthlyTransaction.UserId,
		"monthly_transaction_name":   monthlyTransaction.MonthlyTransactionName,
		"monthly_transaction_amount": monthlyTransaction.MonthlyTransactionAmount,
		"monthly_transaction_date":   monthlyTransaction.MonthlyTransactionDate,
		"category_id":                monthlyTransaction.CategoryId,
		"sub_category_id":            monthlyTransaction.SubCategoryId,
		"include_flg":                true,
		"payment_id":                 paymentId,
	}).Error
}

func (ts *FixedStore) EditFixed(monthlyTransaction *model.EditFixed) error {
	paymentId := interface{}(monthlyTransaction.PaymentId)
	if monthlyTransaction.PaymentId == 0 {
		paymentId = nil
	}

	return ts.db.Table("monthly_transaction").
		Where("monthly_transaction_id = ?", monthlyTransaction.MonthlyTransactionId).
		Where("user_no = ?", monthlyTransaction.UserId).
		Updates(map[string]interface{}{
			"monthly_transaction_name":   monthlyTransaction.MonthlyTransactionName,
			"monthly_transaction_amount": monthlyTransaction.MonthlyTransactionAmount,
			"monthly_transaction_date":   monthlyTransaction.MonthlyTransactionDate,
			"category_id":                monthlyTransaction.CategoryId,
			"sub_category_id":            monthlyTransaction.SubCategoryId,
			"include_flg":                monthlyTransaction.IncludeFlg,
			"payment_id":                 paymentId,
		}).Error
}

func (ts *FixedStore) DeleteFixed(monthlyTransaction *model.DeleteFixed) {
	ts.db.Table("monthly_transaction").
		Where("monthly_transaction_id = ?", monthlyTransaction.MonthlyTransactionId).
		Where("user_no = ?", monthlyTransaction.UserId).
		Delete(&model.DeleteFixed{})
}
