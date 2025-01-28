package store_mysql

import (
	"MoneyHook/MoneyHook-API/model"
	"slices"
	"time"

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
			"t.transaction_date", "c.category_name", "t.category_id", "sc.sub_category_name",
			"t.sub_category_id", "t.fixed_flg", "t.payment_id", "pr.payment_name").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Joins("LEFT JOIN payment_resource pr ON pr.payment_id = t.payment_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ? AND LAST_DAY(?)", month, month).
		Order("t.transaction_date DESC, t.transaction_id DESC").
		Scan(&timeline_list)

	return &timeline_list
}

func search_spending_data(data_list *[]model.MonthlySpendingData, key *string) *model.MonthlySpendingData {
	var result model.MonthlySpendingData
	result.Month = *key

	for _, d := range *data_list {
		if d.Month == *key {
			result.TotalAmount = d.TotalAmount
		}
	}
	return &result
}

func (ts *TransactionStore) GetMonthlySpendingData(userId int, month string) *[]model.MonthlySpendingData {
	var result_list []model.MonthlySpendingData

	var query_list []model.MonthlySpendingData
	ts.db.Unscoped().
		Select("SUM(transaction_amount) as total_amount", "DATE_FORMAT(transaction_date, '%Y-%m-01') as month").
		Table("transaction").
		Where("user_no = ?", userId).
		Where("0 > transaction_amount").
		Where("transaction_date BETWEEN DATE_SUB(?, INTERVAL 5 MONTH) AND LAST_DAY(?)", month, month).
		Group("month").
		Order("month DESC").
		Find(&query_list)

	// 取得できた月のリストを取得
	var query_month_list []string
	for _, q := range query_list {
		query_month_list = append(query_month_list, q.Month)
	}

	// 6ヶ月分のデータを格納
	for i := 0; i < 6; i++ {
		s, _ := time.Parse("2006-01-02", month)
		target_month := s.AddDate(0, -i, 0)
		str_target_month := target_month.Format("2006-01-02")

		if slices.Contains(query_month_list, str_target_month) {
			result_list = append(result_list, *search_spending_data(&query_list, &str_target_month))
			continue
		}

		result_list = append(result_list, model.MonthlySpendingData{TotalAmount: 0, Month: str_target_month})
	}

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

	var amount_condition string
	fixed_condition := ""
	if isSpending {
		amount_condition = "0 > t.transaction_amount"
		fixed_condition = "t.fixed_flg = TRUE"
	} else {
		amount_condition = "0 < t.transaction_amount"
	}

	ts.db.Unscoped().
		Select(
			"c.category_name",
			"sum(t.transaction_amount) OVER(PARTITION BY c.category_name) AS total_category_amount",
			"t.transaction_name",
			"t.transaction_amount",
			"t.transaction_date").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ?", month).
		Where("LAST_DAY(?)", month).
		Where(amount_condition).
		Where(fixed_condition).
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
		"transaction_amount",
		"transaction_date").
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
		"tran_list.transaction_amount",
		"tran_list.transaction_date").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("RIGHT JOIN (?) tran_list ON tran_list.transaction_id = t.transaction_id", subquery_1).
		Joins("RIGHT JOIN (?) sub_clist ON sub_clist.sub_category_id = t.sub_category_id", subquery_2).
		Where("t.user_no = ?", userId).
		Where("0 > t.transaction_amount").
		Where("t.fixed_flg = FALSE").
		Where("t.transaction_date BETWEEN ? AND LAST_DAY(?)", month, month).
		Order("category_total_amount").
		Order("sub_category_total_amount").
		Order("transaction_date").
		Order("transaction_amount").
		Scan(&monthly_variable_data)

	return &monthly_variable_data
}

func (ts *TransactionStore) GetTotalSpending(userId int, categoryId string, subCategoryId string, startMonth string, endMonth string) *[]model.TotalSpendingData {
	var total_spending_data []model.TotalSpendingData

	var condition = make(map[string]interface{})
	if len(categoryId) > 0 {
		condition["c.category_id"] = categoryId
	}
	if len(subCategoryId) > 0 {
		condition["sc.sub_category_id"] = subCategoryId
	}

	subquery_1 := ts.db.Select("transaction_id",
		"transaction_name",
		"transaction_amount").
		Table("transaction").
		Where("user_no = ?", userId).
		Where("transaction_date BETWEEN ? AND LAST_DAY(?)", startMonth, endMonth)

	subquery_2 := ts.db.Select("t.sub_category_id",
		"sc.sub_category_name",
		"SUM(t.transaction_amount) AS sub_category_total_amount").
		Table("transaction t").
		Joins("INNER JOIN sub_category sc ON t.sub_category_id = sc.sub_category_id").
		Where("t.user_no = ?", userId).
		Where("transaction_date BETWEEN ? AND LAST_DAY(?)", startMonth, endMonth).
		Group("t.sub_category_id")

	ts.db.Select("c.category_name",
		"SUM(t.transaction_amount) OVER (PARTITION BY c.category_name) AS category_total_amount",
		"sub_clist.sub_category_id",
		"sub_clist.sub_category_name",
		"sub_clist.sub_category_total_amount",
		"tran_list.transaction_id",
		"tran_list.transaction_name",
		"tran_list.transaction_amount",
		"t.transaction_date").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Joins("RIGHT JOIN (?) tran_list ON tran_list.transaction_id = t.transaction_id", subquery_1).
		Joins("RIGHT JOIN (?) sub_clist ON sub_clist.sub_category_id = t.sub_category_id", subquery_2).
		Where("t.user_no = ?", userId).
		Where("0 > t.transaction_amount").
		Where(condition).
		Where("transaction_date BETWEEN ? AND LAST_DAY(?)", startMonth, endMonth).
		Order("category_total_amount").
		Order("sub_category_total_amount").
		Order("transaction_amount").
		Scan(&total_spending_data)

	return &total_spending_data
}

func (ts *TransactionStore) GetGroupByPayment(userId int, month string) *[]model.PaymentGroupTransaction {
	var payment_group_transaction []model.PaymentGroupTransaction

	ts.db.Select(
		"pr.payment_id",
		"pr.payment_name",
		"SUM(t.transaction_amount) OVER (PARTITION BY pr.payment_name) AS payment_amount",
		"pt.payment_type_id",
		"pt.payment_type_name",
		"pr.payment_date IS NOT NULL AS is_payment_due_later",
		"t.transaction_id",
		"t.transaction_name",
		"t.transaction_amount",
		"t.transaction_date",
		"c.category_id",
		"c.category_name",
		"sc.sub_category_id",
		"sc.sub_category_name",
		"t.fixed_flg").
		Table("transaction t").
		Joins("LEFT JOIN payment_resource pr ON t.payment_id = pr.payment_id").
		Joins("LEFT JOIN payment_type pt ON pr.payment_type_id = pt.payment_type_id").
		Joins("JOIN category c ON c.category_id = t.category_id").
		Joins("JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_amount < 0").
		Where("t.transaction_date BETWEEN ? AND LAST_DAY(?)", month, month).
		Order("payment_amount").
		Order("t.transaction_date DESC").
		Order("t.transaction_id DESC").
		Scan(&payment_group_transaction)

	return &payment_group_transaction
}

func (ts *TransactionStore) GetLastMonthGroupByPayment(userId int, month string) *[]model.PaymentGroupTransaction {
	var payment_group_transaction []model.PaymentGroupTransaction

	ts.db.Select(
		"t.payment_id",
		"SUM(t.transaction_amount) payment_amount").
		Table("transaction t").
		Where("t.user_no = ?", userId).
		Where("t.transaction_amount < 0").
		Where("t.transaction_date BETWEEN ? AND LAST_DAY(?)", month, month).
		Group("t.payment_id").
		Order("payment_amount").
		Scan(&payment_group_transaction)

	return &payment_group_transaction
}

func (ts *TransactionStore) GetMonthlyWithdrawalAmount(userId int, paymentId int, startMonth string, endMonth string) *model.MonthlyWithdrawalAmountList {
	monthlyWithdrawalAmount := model.MonthlyWithdrawalAmountList{AggregationStartDate: startMonth, AggregationEndDate: endMonth}

	ts.db.Select(
		"t.payment_id",
		"pr.payment_name",
		"pr.payment_date",
		"SUM(t.transaction_amount) withdrawal_amount").
		Table("transaction t").
		Joins("LEFT JOIN payment_resource pr ON t.payment_id = pr.payment_id").
		Where("t.user_no = ?", userId).
		Where("t.payment_id = ?", paymentId).
		Where("pr.payment_date IS NOT NULL").
		Where("t.transaction_date BETWEEN ? AND ?", startMonth, endMonth).
		Group("t.payment_id").
		Scan(&monthlyWithdrawalAmount)

	return &monthlyWithdrawalAmount
}

func (ts *TransactionStore) GetFrequentTransactionName(userId int) *[]model.FrequentTransactionName {
	var frequest_transaction_name_list []model.FrequentTransactionName

	ts.db.Table("transaction tran").
		Select("tran.transaction_name",
			"tran.category_id",
			"c.category_name",
			"tran.sub_category_id",
			"sc.sub_category_name",
			"tran.fixed_flg",
			"tran.payment_id").
		Joins("INNER JOIN category c ON tran.category_id = c.category_id").
		Joins("INNER JOIN sub_category sc ON tran.sub_category_id = sc.sub_category_id").
		Where("tran.user_no = ?", userId).
		Group("tran.transaction_name").
		Group("tran.category_id").
		Group("tran.sub_category_id").
		Group("tran.fixed_flg").
		Group("tran.payment_id").
		Order("COUNT(tran.transaction_name) DESC").
		Scan(&frequest_transaction_name_list)

	return &frequest_transaction_name_list
}

func (ts *TransactionStore) AddTransaction(transaction *model.AddTransaction) error {
	paymentId := interface{}(transaction.PaymentId)
	if transaction.PaymentId == 0 {
		paymentId = nil
	}

	return ts.db.Table("transaction").Create(map[string]interface{}{
		"user_no":            transaction.UserId,
		"transaction_name":   transaction.TransactionName,
		"transaction_amount": transaction.TransactionAmount,
		"transaction_date":   transaction.TransactionDate,
		"category_id":        transaction.CategoryId,
		"sub_category_id":    transaction.SubCategoryId,
		"fixed_flg":          transaction.FixedFlg,
		"payment_id":         paymentId,
	}).Error
}

func (ts *TransactionStore) AddTransactionList(transaction *model.AddTransactionList) error {
	var insert_val []map[string]any

	for _, tran := range transaction.TransactionList {
		paymentId := interface{}(tran.PaymentId)
		if tran.PaymentId == 0 {
			paymentId = nil
		}
		insert_val = append(insert_val, map[string]any{
			"user_no":            transaction.UserId,
			"transaction_name":   tran.TransactionName,
			"transaction_amount": tran.TransactionAmount,
			"transaction_date":   tran.TransactionDate,
			"category_id":        tran.CategoryId,
			"sub_category_id":    tran.SubCategoryId,
			"fixed_flg":          tran.FixedFlg,
			"payment_id":         paymentId,
		})
	}

	return ts.db.Table("transaction").
		Create(insert_val).Error
}

func (ts *TransactionStore) EditTransaction(transaction *model.EditTransaction) error {
	paymentId := interface{}(transaction.PaymentId)
	if transaction.PaymentId == 0 {
		paymentId = nil
	}

	return ts.db.Table("transaction").
		Where("transaction_id = ?", transaction.TransactionId).
		Where("user_no = ?", transaction.UserId).
		Updates(map[string]interface{}{
			"transaction_name":   transaction.TransactionName,
			"transaction_amount": transaction.TransactionAmount,
			"transaction_date":   transaction.TransactionDate,
			"category_id":        transaction.CategoryId,
			"sub_category_id":    transaction.SubCategoryId,
			"fixed_flg":          transaction.FixedFlg,
			"payment_id":         paymentId,
		}).Error
}

func (ts *TransactionStore) DeleteTransaction(transaction *model.DeleteTransaction) error {
	return ts.db.Table("transaction").
		Where("transaction_id = ?", transaction.TransactionId).
		Where("user_no = ?", transaction.UserId).
		Delete(&model.DeleteTransaction{}).Error
}
