package store_postgres

import (
	"MoneyHook/MoneyHook-API/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

type TransactionStore struct {
	db *gorm.DB
}

func NewTransactionStore(db *gorm.DB) *TransactionStore {
	return &TransactionStore{db: db}
}

func (ts *TransactionStore) GetTimelineData(userId string, month string) (*[]model.Timeline, error) {
	var timeline_list []model.Timeline

	err := ts.db.Unscoped().Preload("Category").
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
		Where("t.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", month, month).
		Order("t.transaction_date DESC, t.transaction_id DESC").
		Scan(&timeline_list).Error

	return &timeline_list, err
}

func (ts *TransactionStore) GetMonthlySpendingData(userId string, month string) (*[]model.MonthlySpendingData, error) {
	var query_list []model.MonthlySpendingData
	err := ts.db.Unscoped().
		Select("SUM(transaction_amount) as total_amount",
			"TO_CHAR(date_trunc('month', transaction_date), 'YYYY-MM-DD') as month").
		Table("transaction").
		Where("user_no = ?", userId).
		Where("transaction_amount < 0").
		Where("transaction_date BETWEEN (?::date - INTERVAL '5 months') AND (date_trunc('month', ?::date) + INTERVAL '1 month' - INTERVAL '1 day')", month, month).
		Group("month").
		Order("month DESC").
		Find(&query_list).Error

	if err != nil {
		return nil, err
	}
	return fillMonthlySpendingData(query_list, month)
}

func fillMonthlySpendingData(rows []model.MonthlySpendingData, month string) (*[]model.MonthlySpendingData, error) {
	start, err := time.Parse("2006-01-02", month)
	if err != nil {
		return nil, err
	}
	amounts := make(map[string]int, len(rows))
	for _, row := range rows {
		amounts[row.Month] = row.TotalAmount
	}
	result := make([]model.MonthlySpendingData, 0, 6)
	for i := 0; i < 6; i++ {
		month := start.AddDate(0, -i, 0).Format("2006-01-02")
		result = append(result, model.MonthlySpendingData{Month: month, TotalAmount: amounts[month]})
	}
	return &result, nil
}

func (ts *TransactionStore) GetTransactionData(userId string, transactionId string) (*model.TransactionData, error) {
	var result model.TransactionData

	tx := ts.db.Unscoped().
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
		Take(&result)

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}

	return &result, nil
}

func (ts *TransactionStore) GetMonthlyFixedData(userId string, month string, isSpending bool) (*[]model.MonthlyFixedData, error) {
	var result_list []model.MonthlyFixedData

	var amount_condition string
	if isSpending {
		amount_condition = "t.transaction_amount < 0"
	} else {
		amount_condition = "t.transaction_amount > 0"
	}

	query := ts.db.Unscoped().
		Select(
			"c.category_id",
			"c.category_name",
			"SUM(t.transaction_amount) OVER (PARTITION BY c.category_name) AS total_category_amount",
			"sc.sub_category_id",
			"sc.sub_category_name",
			"t.transaction_id",
			"t.transaction_name",
			"t.transaction_amount",
			"t.transaction_date",
			"t.fixed_flg",
			"pr.payment_id",
			"pr.payment_name").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Joins("LEFT OUTER JOIN payment_resource pr ON t.payment_id = pr.payment_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", month, month).
		Where(amount_condition)

	if isSpending {
		query = query.Where("t.fixed_flg = TRUE")
	}

	err := query.
		Order("total_category_amount, transaction_amount").
		Find(&result_list).Error

	return &result_list, err
}

func (ts *TransactionStore) GetHome(userId string, month string) (*[]model.HomeCategory, error) {
	var home_data []model.HomeCategory

	subquery := ts.db.Select("st.sub_category_id",
		"ssc.category_id",
		"ssc.sub_category_name",
		"SUM(st.transaction_amount) AS sub_category_total_amount").
		Table("transaction AS st").
		Joins("INNER JOIN sub_category AS ssc ON ssc.sub_category_id = st.sub_category_id").
		Where("st.user_no = ?", userId).
		Where("st.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", month, month).
		Group("st.sub_category_id, ssc.category_id, ssc.sub_category_name")

	err := ts.db.Select("sub_tran.category_id",
		"(SELECT category_name FROM category WHERE category_id = sub_tran.category_id) AS category_name",
		"SUM(sub_tran.sub_category_total_amount) OVER (PARTITION BY sub_tran.category_id) AS category_total_amount",
		"sub_tran.category_id AS category_id_02",
		"sub_tran.sub_category_id",
		"sub_tran.sub_category_name",
		"sub_tran.sub_category_total_amount").
		Table("transaction AS t").
		Joins("RIGHT JOIN (?) AS sub_tran ON sub_tran.sub_category_id = t.sub_category_id", subquery).
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", month, month).
		Where("t.transaction_amount < 0").
		Group("sub_tran.category_id, sub_tran.sub_category_id, sub_tran.sub_category_name, sub_tran.sub_category_total_amount").
		Order("category_total_amount, sub_tran.sub_category_id").
		Scan(&home_data).Error

	return &home_data, err
}

func (ts *TransactionStore) GetMonthlyVariableData(userId string, month string) (*[]model.MonthlyVariableData, error) {
	var monthly_variable_data []model.MonthlyVariableData

	subquery_1 := ts.db.Select("transaction_id",
		"transaction_name",
		"transaction_amount",
		"transaction_date",
		"payment_id").
		Table("transaction").
		Where("user_no = ?", userId).
		Where("transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", month, month)

	subquery_2 := ts.db.Select("t.sub_category_id",
		"sc.sub_category_name",
		"SUM(t.transaction_amount) AS sub_category_total_amount").
		Table("transaction t").
		Joins("INNER JOIN sub_category sc ON t.sub_category_id = sc.sub_category_id").
		Where("t.user_no = ?", userId).
		Where("t.fixed_flg = FALSE").
		Where("t.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", month, month).
		Group("t.sub_category_id, sc.sub_category_name")

	err := ts.db.Select("c.category_id",
		"c.category_name",
		"SUM(t.transaction_amount) OVER (PARTITION BY c.category_name) AS category_total_amount",
		"sub_clist.sub_category_id",
		"sub_clist.sub_category_name",
		"sub_clist.sub_category_total_amount",
		"tran_list.transaction_id",
		"tran_list.transaction_name",
		"tran_list.transaction_amount",
		"tran_list.transaction_date",
		"pr.payment_id",
		"pr.payment_name").
		Table("transaction t").
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("RIGHT JOIN (?) AS tran_list ON tran_list.transaction_id = t.transaction_id", subquery_1).
		Joins("RIGHT JOIN (?) AS sub_clist ON sub_clist.sub_category_id = t.sub_category_id", subquery_2).
		Joins("LEFT OUTER JOIN payment_resource pr ON tran_list.payment_id = pr.payment_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_amount < 0").
		Where("t.fixed_flg = FALSE").
		Where("t.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", month, month).
		Order("category_total_amount").
		Order("sub_category_total_amount").
		Order("tran_list.transaction_date").
		Order("tran_list.transaction_amount").
		Scan(&monthly_variable_data).Error

	return &monthly_variable_data, err
}

func (ts *TransactionStore) GetTotalSpending(userId string, categoryId string, subCategoryId string, startMonth string, endMonth string) (*[]model.TotalSpendingData, error) {
	var total_spending_data []model.TotalSpendingData

	query := ts.db

	if len(categoryId) > 0 {
		query = query.Where("c.category_id = ?", categoryId)
	}
	if len(subCategoryId) > 0 {
		query = query.Where("sc.sub_category_id = ?", subCategoryId)
	}

	subquery_1 := ts.db.Select(
		"transaction_id",
		"transaction_name",
		"transaction_amount").
		Table("transaction").
		Where("user_no = ?", userId).
		Where("transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", startMonth, endMonth)

	subquery_2 := ts.db.Select(
		"t.sub_category_id",
		"sc.sub_category_name",
		"SUM(t.transaction_amount) AS sub_category_total_amount").
		Table("transaction t").
		Joins("INNER JOIN sub_category sc ON t.sub_category_id = sc.sub_category_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", startMonth, endMonth).
		Group("t.sub_category_id, sc.sub_category_name")

	err := query.Select(
		"c.category_name",
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
		Joins("RIGHT JOIN (?) AS tran_list ON tran_list.transaction_id = t.transaction_id", subquery_1).
		Joins("RIGHT JOIN (?) AS sub_clist ON sub_clist.sub_category_id = t.sub_category_id", subquery_2).
		Where("t.user_no = ?", userId).
		Where("t.transaction_amount < 0").
		Where("t.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", startMonth, endMonth).
		Order("category_total_amount").
		Order("sub_category_total_amount").
		Order("tran_list.transaction_amount").
		Scan(&total_spending_data).Error

	return &total_spending_data, err
}

func (ts *TransactionStore) GetGroupByPayment(userId string, month string) (*[]model.PaymentGroupTransaction, error) {
	var payment_group_transaction []model.PaymentGroupTransaction

	err := ts.db.Select(
		"pr.payment_id",
		"pr.payment_name",
		"SUM(t.transaction_amount) OVER (PARTITION BY pr.payment_name) AS payment_amount",
		"pt.payment_type_id",
		"pt.payment_type_name",
		"(pr.payment_date IS NOT NULL) AS is_payment_due_later",
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
		Where("t.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", month, month).
		Order("payment_amount").
		Order("t.transaction_date DESC").
		Order("t.transaction_id DESC").
		Scan(&payment_group_transaction).Error

	return &payment_group_transaction, err
}

func (ts *TransactionStore) GetLastMonthGroupByPayment(userId string, month string) (*[]model.PaymentGroupTransaction, error) {
	var payment_group_transaction []model.PaymentGroupTransaction

	err := ts.db.Select(
		"t.payment_id",
		"SUM(t.transaction_amount) AS payment_amount").
		Table("transaction t").
		Where("t.user_no = ?", userId).
		Where("t.transaction_amount < 0").
		Where("t.transaction_date BETWEEN ? AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')", month, month).
		Group("t.payment_id").
		Order("payment_amount").
		Scan(&payment_group_transaction).Error

	return &payment_group_transaction, err
}

func (ts *TransactionStore) GetMonthlyWithdrawalAmount(userId string, paymentId string, startMonth string, endMonth string) (*model.MonthlyWithdrawalAmountList, error) {
	var monthlyWithdrawalAmount model.MonthlyWithdrawalAmountList

	err := ts.db.Select(
		"t.payment_id",
		"pr.payment_name",
		"pr.payment_date",
		"SUM(t.transaction_amount) AS withdrawal_amount").
		Table("transaction t").
		Joins("LEFT JOIN payment_resource pr ON t.payment_id = pr.payment_id").
		Where("t.user_no = ?", userId).
		Where("t.payment_id = ?", paymentId).
		Where("pr.payment_date IS NOT NULL").
		Where(`
			t.transaction_date BETWEEN ?::date 
			AND (date_trunc('month', ?::date) + interval '1 month' - interval '1 day')
		`, startMonth, endMonth).
		Group("t.payment_id, pr.payment_name, pr.payment_date").
		Scan(&monthlyWithdrawalAmount).Error

	monthlyWithdrawalAmount.AggregationStartDate = startMonth
	monthlyWithdrawalAmount.AggregationEndDate = endMonth

	return &monthlyWithdrawalAmount, err
}

func (ts *TransactionStore) GetFrequentTransactionName(userId string, limit int) (*[]model.FrequentTransactionName, error) {
	var frequent_transaction_name_list []model.FrequentTransactionName

	err := frequentTransactionNameQuery(ts.db, userId, limit).Scan(&frequent_transaction_name_list).Error

	return &frequent_transaction_name_list, err
}

func frequentTransactionNameQuery(db *gorm.DB, userId string, limit int) *gorm.DB {
	frequentTransactions := db.Table("transaction tran").
		Select("tran.transaction_name",
			"tran.category_id",
			"c.category_name",
			"tran.sub_category_id",
			"sc.sub_category_name",
			"tran.fixed_flg",
			"tran.payment_id",
			"COUNT(*) AS usage_count",
			"ROW_NUMBER() OVER (PARTITION BY tran.transaction_name ORDER BY COUNT(*) DESC, tran.category_id ASC, tran.sub_category_id ASC, tran.fixed_flg ASC, tran.payment_id ASC NULLS FIRST) AS row_num").
		Joins("INNER JOIN category c ON tran.category_id = c.category_id").
		Joins("INNER JOIN sub_category sc ON tran.sub_category_id = sc.sub_category_id").
		Where("tran.user_no = ?", userId).
		Group("tran.transaction_name, tran.category_id, c.category_name, tran.sub_category_id, sc.sub_category_name, tran.fixed_flg, tran.payment_id").
		Order("usage_count DESC, tran.transaction_name ASC, tran.category_id ASC, tran.sub_category_id ASC, tran.fixed_flg ASC, tran.payment_id ASC NULLS FIRST")

	return db.Table("(?) AS frequent_transactions", frequentTransactions).
		Where("row_num = ?", 1).
		Order("usage_count DESC, transaction_name ASC").
		Limit(limit)
}

func (ts *TransactionStore) AddTransaction(transaction *model.AddTransaction) error {
	return ts.db.Transaction(func(tx *gorm.DB) error {
		input := *transaction
		transaction = &input
		subCategoryID, err := resolveWriteSubCategory(tx, transaction.UserId, transaction.CategoryId, transaction.SubCategoryId, transaction.SubCategoryName)
		if err != nil {
			return err
		}
		transaction.SubCategoryId = subCategoryID

		paymentId := interface{}(transaction.PaymentId)
		if transaction.PaymentId == "" {
			paymentId = nil
		}

		return tx.Table("transaction").Create(map[string]interface{}{
			"user_no":            transaction.UserId,
			"transaction_name":   transaction.TransactionName,
			"transaction_amount": transaction.TransactionAmount,
			"transaction_date":   transaction.TransactionDate,
			"category_id":        transaction.CategoryId,
			"sub_category_id":    transaction.SubCategoryId,
			"fixed_flg":          transaction.FixedFlg,
			"payment_id":         paymentId,
		}).Error

	})
}

func (ts *TransactionStore) AddTransactionList(transaction *model.AddTransactionList) error {
	return ts.db.Transaction(func(tx *gorm.DB) error {
		if len(transaction.TransactionList) == 0 {
			return errors.New("transaction list is empty")
		}
		var insert_val []map[string]any

		for _, tran := range transaction.TransactionList {
			subCategoryID, err := resolveWriteSubCategory(tx, transaction.UserId, tran.CategoryId, tran.SubCategoryId, tran.SubCategoryName)
			if err != nil {
				return err
			}
			tran.SubCategoryId = subCategoryID
			paymentId := interface{}(tran.PaymentId)
			if tran.PaymentId == "" {
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

		return tx.Table("transaction").
			Create(insert_val).Error

	})
}

func (ts *TransactionStore) EditTransaction(transaction *model.EditTransaction) error {
	return ts.db.Transaction(func(tx *gorm.DB) error {
		input := *transaction
		transaction = &input
		subCategoryID, err := resolveWriteSubCategory(tx, transaction.UserId, transaction.CategoryId, transaction.SubCategoryId, transaction.SubCategoryName)
		if err != nil {
			return err
		}
		transaction.SubCategoryId = subCategoryID

		paymentId := interface{}(transaction.PaymentId)
		if transaction.PaymentId == "" {
			paymentId = nil
		}

		result := tx.Table("transaction").
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
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil

	})
}

func (ts *TransactionStore) DeleteTransaction(transaction *model.DeleteTransaction) error {
	return ts.db.Table("transaction").
		Where("transaction_id = ?", transaction.TransactionId).
		Where("user_no = ?", transaction.UserId).
		Delete(&model.DeleteTransaction{}).Error
}
