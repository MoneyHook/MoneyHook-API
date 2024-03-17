package store

import (
	"MoneyHook/MoneyHook-API/model"

	"gorm.io/gorm"
)

type TransactionStore struct {
	db *gorm.DB
}

func NewTransactionStore(db *gorm.DB) *TransactionStore {
	return &TransactionStore{db: db}
}

func (cs *TransactionStore) GetTimelineData(userId int, month string) *[]model.Timeline {
	var timeline_list []model.Timeline

	cs.db.Unscoped().Preload("Category").
		Preload("SubCategory").
		Select("t.transaction_id", "t.transaction_name", "ABS(t.transaction_amount) AS transaction_amount",
			"CASE WHEN t.transaction_amount > 0 THEN 1 ELSE -1 END AS transaction_sign",
			"t.transaction_date", "c.category_name", "t.category_id", "sc.sub_category_name", "t.sub_category_id", "t.fixed_flg").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ? AND LAST_DAY(?)", month, month).
		Order("t.transaction_date DESC, t.transaction_id DESC").
		Debug().
		Find(&timeline_list)

	return &timeline_list
}
func (cs *TransactionStore) GetMonthlySpendingData(userId int, month string) *[]model.MonthlySpendingData {
	var result_list []model.MonthlySpendingData

	cs.db.Unscoped().
		Select("SUM(transaction_amount) as total_amount", "DATE_FORMAT(transaction_date, '%Y-%m-01') as month").
		Table("transaction").
		Where("user_no = ?", userId).
		Where("0 > transaction_amount").
		Where("transaction_date BETWEEN DATE_SUB(?, INTERVAL 5 MONTH) AND LAST_DAY(?)", month, month).
		Group("month").
		Order("month DESC").
		Find(&result_list)

	return &result_list
}

func (cs *TransactionStore) GetTransactionData(userId int, transactionId int) *model.TransactionData {
	var result model.TransactionData

	cs.db.Unscoped().
		Select(
			"t.transaction_date",
			"t.transaction_name",
			"t.transaction_amount",
			"t.category_id",
			"c.category_name",
			"t.sub_category_id",
			"sc.sub_category_name",
			"t.fixed_flg").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_id = ?", transactionId).
		Find(&result)

	return &result
}

func (cs *TransactionStore) GetMonthlyFixedIncomeData(userId int, month string) *model.MonthlyFixedIncome {
	var result model.MonthlyFixedIncome

	cs.db.Unscoped().
		Select(
			"c.category_id",
			"c.category_name",
			"sum(t.transaction_amount) OVER(PARTITION BY c.category_name) AS total_category_amount",
			"t.category_id AS transaction_category_id",
			"t.transaction_name",
			"t.transaction_amount"
			).
		Table("transaction t").
		Joins(INNER JOIN category c ON c.category_id = t.category_id).
		Joins(INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id).
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ?", month).
		Where("LAST_DAY(?)", month).
		Where("0 > t.transaction_amount").
		Where(t.fixed_flg = TRUE).
		Order(total_category_amount).
		Debug().
		Find(&result)
	
	// for _,v :=range result{
	// 	fmt.Println(v)
	// }

	return &result
}
