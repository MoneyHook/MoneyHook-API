package store_postgres

import (
	"MoneyHook/MoneyHook-API/model"

	"gorm.io/gorm"
)

type JobStore struct {
	db *gorm.DB
}

func NewJobStore(db *gorm.DB) *JobStore {
	return &JobStore{db: db}
}

func (js *JobStore) SelectMonthlyTransaction(date int, isLastDay bool) (*[]model.JobMonthlyTransaction, error) {
	var fixed_list []model.JobMonthlyTransaction

	var mt_date_condition string
	if isLastDay {
		mt_date_condition = "? <= monthly_transaction_date"
	} else {
		mt_date_condition = "? = monthly_transaction_date"
	}

	err := js.db.Unscoped().
		Select("monthly_transaction_id",
			"user_no",
			"monthly_transaction_name",
			"monthly_transaction_amount",
			"monthly_transaction_date",
			"category_id",
			"sub_category_id",
			"payment_id",
		).
		Table("monthly_transaction").
		Where(mt_date_condition, date).
		Where("include_flg = TRUE").
		Find(&fixed_list).Error

	return &fixed_list, err
}

func (js *JobStore) InsertTransaction(transactions *[]model.JobTransaction) error {
	var insert_val []map[string]any

	for _, tran := range *transactions {
		paymentId := interface{}(tran.PaymentId)
		if tran.PaymentId == "" {
			paymentId = nil
		}
		insert_val = append(insert_val, map[string]any{
			"user_no":            tran.UserNo,
			"transaction_name":   tran.TransactionName,
			"transaction_amount": tran.TransactionAmount,
			"transaction_date":   tran.TransactionDate,
			"category_id":        tran.CategoryId,
			"sub_category_id":    tran.SubCategoryId,
			"fixed_flg":          tran.FixedFlg,
			"payment_id":         paymentId,
		})
	}

	return js.db.Table("transaction").
		Create(insert_val).Error
}
