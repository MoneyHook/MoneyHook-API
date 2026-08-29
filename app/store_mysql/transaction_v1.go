package store_mysql

import (
	"MoneyHook/MoneyHook-API/model"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
	"errors"
	"strconv"

	"gorm.io/gorm"
)

type v1TransactionRecord struct {
	TransactionId     uint64  `gorm:"column:transaction_id;primaryKey;autoIncrement"`
	UserId            string  `gorm:"column:user_no"`
	TransactionName   string  `gorm:"column:transaction_name"`
	TransactionAmount int64   `gorm:"column:transaction_amount"`
	TransactionDate   string  `gorm:"column:transaction_date"`
	TransactionTime   *string `gorm:"column:transaction_time"`
	CategoryId        string  `gorm:"column:category_id"`
	SubCategoryId     string  `gorm:"column:sub_category_id"`
	FixedFlg          bool    `gorm:"column:fixed_flg"`
	PaymentId         *string `gorm:"column:payment_id"`
}

func (v1TransactionRecord) TableName() string { return "transaction" }

func (ts *TransactionStore) GetV1Transaction(userId string, transactionId string) (*model.V1Transaction, error) {
	return getMySQLV1Transaction(ts.db, userId, transactionId)
}

func getMySQLV1Transaction(db *gorm.DB, userId string, transactionId string) (*model.V1Transaction, error) {
	var result model.V1Transaction
	err := db.Table("transaction t").
		Select(
			"CAST(t.transaction_id AS CHAR) AS transaction_id",
			"DATE_FORMAT(t.transaction_date, '%Y-%m-%d') AS transaction_date",
			"TIME_FORMAT(t.transaction_time, '%H:%i') AS transaction_time",
			"t.transaction_name",
			"ABS(t.transaction_amount) AS amount",
			"CASE WHEN t.transaction_amount > 0 THEN 1 ELSE -1 END AS sign",
			"t.transaction_amount AS signed_amount",
			"CAST(t.category_id AS CHAR) AS category_id",
			"c.category_name",
			"CAST(t.sub_category_id AS CHAR) AS sub_category_id",
			"sc.sub_category_name",
			"t.fixed_flg",
			"CAST(t.payment_id AS CHAR) AS payment_id",
			"pr.payment_name",
		).
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Joins("LEFT JOIN payment_resource pr ON pr.payment_id = t.payment_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_id = ?", transactionId).
		Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, transactiondomain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (ts *TransactionStore) CreateV1Transaction(input *model.V1TransactionWrite) (*model.V1Transaction, error) {
	var created *model.V1Transaction
	err := ts.db.Transaction(func(tx *gorm.DB) error {
		if err := validateMySQLV1Relations(tx, input); err != nil {
			return err
		}
		record := v1TransactionRecord{
			UserId:            input.UserId,
			TransactionName:   input.TransactionName,
			TransactionAmount: input.Amount * int64(input.Sign),
			TransactionDate:   input.TransactionDate,
			TransactionTime:   input.TransactionTime,
			CategoryId:        input.CategoryId,
			SubCategoryId:     input.SubCategoryId,
			FixedFlg:          input.FixedFlg,
			PaymentId:         input.PaymentId,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		var err error
		created, err = getMySQLV1Transaction(tx, input.UserId, strconv.FormatUint(record.TransactionId, 10))
		return err
	})
	return created, err
}

func (ts *TransactionStore) UpdateV1Transaction(input *model.V1TransactionWrite) (*model.V1Transaction, string, error) {
	var updated *model.V1Transaction
	var previousDate string
	err := ts.db.Transaction(func(tx *gorm.DB) error {
		current, err := getMySQLV1Transaction(tx, input.UserId, input.TransactionId)
		if err != nil {
			return err
		}
		previousDate = current.TransactionDate
		if err := validateMySQLV1Relations(tx, input); err != nil {
			return err
		}
		result := tx.Table("transaction").
			Where("transaction_id = ?", input.TransactionId).
			Where("user_no = ?", input.UserId).
			Updates(map[string]any{
				"transaction_name":   input.TransactionName,
				"transaction_amount": input.Amount * int64(input.Sign),
				"transaction_date":   input.TransactionDate,
				"transaction_time":   input.TransactionTime,
				"category_id":        input.CategoryId,
				"sub_category_id":    input.SubCategoryId,
				"fixed_flg":          input.FixedFlg,
				"payment_id":         input.PaymentId,
			})
		if result.Error != nil {
			return result.Error
		}
		updated, err = getMySQLV1Transaction(tx, input.UserId, input.TransactionId)
		return err
	})
	return updated, previousDate, err
}

func (ts *TransactionStore) DeleteV1Transaction(userId string, transactionId string) error {
	result := ts.db.Table("transaction").
		Where("transaction_id = ?", transactionId).
		Where("user_no = ?", userId).
		Delete(&v1TransactionRecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return transactiondomain.ErrNotFound
	}
	return nil
}

func (ts *TransactionStore) GetV1AnalyticsTransactions(userId string, startDate string, endDate string) ([]model.V1AnalyticsTransaction, error) {
	result := make([]model.V1AnalyticsTransaction, 0)
	err := ts.db.Table("transaction t").
		Select(
			"CAST(t.transaction_id AS CHAR) AS transaction_id",
			"DATE_FORMAT(t.transaction_date, '%Y-%m-%d') AS transaction_date",
			"TIME_FORMAT(t.transaction_time, '%H:%i') AS transaction_time",
			"t.transaction_name",
			"t.transaction_amount AS signed_amount",
			"CAST(t.category_id AS CHAR) AS category_id",
			"c.category_name",
			"CAST(t.sub_category_id AS CHAR) AS sub_category_id",
			"sc.sub_category_name",
			"t.fixed_flg",
			"CAST(t.payment_id AS CHAR) AS payment_id",
			"pr.payment_name",
			"CAST(pr.payment_type_id AS CHAR) AS payment_type_id",
			"pt.payment_type_name",
			"pt.is_payment_due_later",
		).
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Joins("LEFT JOIN payment_resource pr ON pr.payment_id = t.payment_id").
		Joins("LEFT JOIN payment_type pt ON pt.payment_type_id = pr.payment_type_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ? AND ?", startDate, endDate).
		Order("t.transaction_date DESC, t.transaction_time IS NULL, t.transaction_time DESC, t.transaction_id DESC").
		Scan(&result).Error
	return result, err
}

func validateMySQLV1Relations(db *gorm.DB, input *model.V1TransactionWrite) error {
	var subCategoryCount int64
	if err := db.Table("sub_category").
		Where("sub_category_id = ?", input.SubCategoryId).
		Where("category_id = ?", input.CategoryId).
		Where("user_no = ? OR user_no = ?", input.UserId, 1).
		Where("NOT EXISTS (SELECT 1 FROM hidden_sub_category hsc WHERE hsc.sub_category_id = sub_category.sub_category_id AND hsc.user_no = ?)", input.UserId).
		Count(&subCategoryCount).Error; err != nil {
		return err
	}
	if subCategoryCount != 1 {
		return transactiondomain.ErrInvalidRelation
	}
	if input.PaymentId == nil {
		return nil
	}
	var paymentCount int64
	if err := db.Table("payment_resource").
		Where("payment_id = ?", *input.PaymentId).
		Where("user_no = ?", input.UserId).
		Count(&paymentCount).Error; err != nil {
		return err
	}
	if paymentCount != 1 {
		return transactiondomain.ErrInvalidRelation
	}
	return nil
}
