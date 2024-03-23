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

func (ts *TransactionStore) GetTimelineData(userId int, month string) *[]model.Timeline {
	var timeline_list []model.Timeline

	ts.db.Unscoped().Preload("Category").
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

func (ts *TransactionStore) GetMonthlySpendingData(userId int, month string) *[]model.MonthlySpendingData {
	var result_list []model.MonthlySpendingData

	ts.db.Unscoped().
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

func (ts *TransactionStore) GetTransactionData(userId int, transactionId int) *model.TransactionData {
	var result model.TransactionData

	ts.db.Unscoped().
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

func (ts *TransactionStore) GetMonthlyFixedData(userId int, month string, isSpending bool) *[]model.MonthlyFixedData {
	var result_list []model.MonthlyFixedData

	var conditions string
	if isSpending {
		conditions = "0 > t.transaction_amount"
	} else {
		conditions = "0 < t.transaction_amount"
	}

	ts.db.Unscoped().
		Select(
			"c.category_name",
			"sum(t.transaction_amount) OVER(PARTITION BY c.category_name) AS total_category_amount",
			"t.transaction_name",
			"t.transaction_amount").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ?", month).
		Where("LAST_DAY(?)", month).
		Where(conditions).
		Where("t.fixed_flg = TRUE").
		Order("total_category_amount").
		Find(&result_list)

	return &result_list
}

func (ts *TransactionStore) GetHome(userId int, month string) *[]model.HomeCategory {
	var home_data []model.HomeCategory

	subquery := ts.db.Select("st.sub_category_id",
		"ssc.category_id",
		"ssc.sub_category_name",
		"SUM(st.transaction_amount) AS sub_category_total_amount").
		Table("transaction AS st").
		Joins("INNER JOIN sub_category AS ssc ON ssc.sub_category_id = st.sub_category_id").
		Where("st.user_no = ?", userId).
		Where("st.transaction_date BETWEEN ? AND LAST_DAY(?)", month, month).
		Group("st.sub_category_id")

	ts.db.Select("sub_tran.category_id",
		"(select category_name from category where category_id = sub_tran.category_id) AS category_name",
		"SUM(sub_tran.sub_category_total_amount) OVER (PARTITION BY sub_tran.category_id) AS category_total_amount",
		"sub_tran.category_id AS category_id_02",
		"sub_tran.sub_category_id",
		"sub_tran.sub_category_name",
		"sub_tran.sub_category_total_amount").
		Table("transaction AS t").
		Joins("RIGHT JOIN (?) sub_tran ON sub_tran.sub_category_id = t.sub_category_id", subquery).
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ? AND LAST_DAY(?)", month, month).
		Where("t.transaction_amount < 0").
		Group("t.sub_category_id").
		Order("category_total_amount, t.sub_category_id").
		Scan(&home_data)

	return &home_data
}

func (ts *TransactionStore) GetMonthlyVariableData(userId int, month string) *[]model.MonthlyVariableData {
	var monthly_variable_data []model.MonthlyVariableData

	subquery_1 := ts.db.Select("transaction_id",
		"transaction_name",
		"transaction_amount").
		Table("transaction").
		Where("user_no = ?", userId).
		Where("transaction_date BETWEEN ? AND LAST_DAY(?)", month, month)

	subquery_2 := ts.db.Select("t.sub_category_id",
		"sc.sub_category_name",
		"SUM(t.transaction_amount) AS sub_category_total_amount").
		Table("transaction t").
		Joins("INNER JOIN sub_category sc ON t.sub_category_id = sc.sub_category_id").
		Where("t.user_no = ?", userId).
		Where("t.fixed_flg = FALSE").
		Where("transaction_date BETWEEN ? AND LAST_DAY(?)", month, month).
		Group("t.sub_category_id")

	ts.db.Select("c.category_name",
		"SUM(t.transaction_amount) OVER (PARTITION BY c.category_name) AS category_total_amount",
		"sub_clist.sub_category_id",
		"sub_clist.sub_category_name",
		"sub_clist.sub_category_total_amount",
		"tran_list.transaction_id",
		"tran_list.transaction_name",
		"tran_list.transaction_amount").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("RIGHT JOIN (?) tran_list ON tran_list.transaction_id = t.transaction_id", subquery_1).
		Joins("RIGHT JOIN (?) sub_clist ON sub_clist.sub_category_id = t.sub_category_id", subquery_2).
		Where("t.user_no = ?", userId).
		Where("0 > t.transaction_amount").
		Where("t.fixed_flg = FALSE").
		Where("transaction_date BETWEEN ? AND LAST_DAY  (?)", month, month).
		Order("category_total_amount").
		Order("sub_category_total_amount").
		Order("transaction_amount").
		Scan(&monthly_variable_data)

	return &monthly_variable_data
}
