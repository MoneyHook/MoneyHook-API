package store_postgres

import (
	household "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/model"
	subcategorydomain "MoneyHook/MoneyHook-API/subcategory"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const v1TransactionTimeSelect = "LEFT(CAST(t.transaction_time AS TEXT), 5) AS transaction_time"

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
	return getPostgresV1Transaction(ts.db, userId, transactionId)
}

func getPostgresV1Transaction(db *gorm.DB, userId string, transactionId string) (*model.V1Transaction, error) {
	var result model.V1Transaction
	err := db.Table("transaction t").Where("t.deleted_at IS NULL").
		Select(
			"CAST(t.transaction_id AS TEXT) AS transaction_id",
			"TO_CHAR(t.transaction_date, 'YYYY-MM-DD') AS transaction_date",
			v1TransactionTimeSelect,
			"t.transaction_name",
			"t.version",
			"EXISTS (SELECT 1 FROM household_entry he WHERE he.source_transaction_id=t.transaction_id AND he.kind='shared' AND he.state='active') AS shared",
			"ABS(t.transaction_amount) AS amount",
			"CASE WHEN t.transaction_amount > 0 THEN 1 ELSE -1 END AS sign",
			"t.transaction_amount AS signed_amount",
			"CAST(t.category_id AS TEXT) AS category_id",
			"c.category_name",
			"CAST(t.sub_category_id AS TEXT) AS sub_category_id",
			"sc.sub_category_name",
			"t.fixed_flg",
			"CAST(t.payment_id AS TEXT) AS payment_id",
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
	err := householdTransaction(ts.db, func(tx *gorm.DB) error {
		if err := validatePostgresV1BaseRelations(tx, input); err != nil {
			return err
		}
		resolvedInput := *input
		if resolvedInput.SubCategoryId == "" {
			subCategoryID, err := resolveV1WriteSubCategory(tx, &resolvedInput)
			if err != nil {
				return err
			}
			resolvedInput.SubCategoryId = subCategoryID
		}
		if err := validatePostgresV1SubCategory(tx, &resolvedInput); err != nil {
			return err
		}
		record := v1TransactionRecord{
			UserId:            resolvedInput.UserId,
			TransactionName:   resolvedInput.TransactionName,
			TransactionAmount: resolvedInput.Amount * int64(resolvedInput.Sign),
			TransactionDate:   resolvedInput.TransactionDate,
			TransactionTime:   resolvedInput.TransactionTime,
			CategoryId:        resolvedInput.CategoryId,
			SubCategoryId:     resolvedInput.SubCategoryId,
			FixedFlg:          resolvedInput.FixedFlg,
			PaymentId:         resolvedInput.PaymentId,
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		var err error
		created, err = getPostgresV1Transaction(tx, resolvedInput.UserId, stringID(record.TransactionId))
		return err
	})
	return created, err
}

func (ts *TransactionStore) UpdateV1Transaction(input *model.V1TransactionWrite) (*model.V1Transaction, string, error) {
	var updated *model.V1Transaction
	var previousDate string
	err := householdTransaction(ts.db, func(tx *gorm.DB) error {
		changes, err := beginHouseholdSourceChange(tx, input.UserId, input.TransactionId)
		if err != nil {
			return err
		}
		current, err := getPostgresV1Transaction(tx, input.UserId, input.TransactionId)
		if err != nil {
			return err
		}
		if input.ExpectedVersion != nil && *input.ExpectedVersion != current.Version {
			return household.Conflict
		}
		previousDate = current.TransactionDate
		if err := validatePostgresV1BaseRelations(tx, input); err != nil {
			return err
		}
		if err := validatePostgresV1SubCategory(tx, input); err != nil {
			return err
		}
		result := tx.Table("transaction").
			Where("transaction_id = ?", input.TransactionId).
			Where("user_no = ?", input.UserId).
			Updates(map[string]any{
				"version":            gorm.Expr("version + 1"),
				"updated_at":         gorm.Expr("CURRENT_TIMESTAMP"),
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
		if err := finishHouseholdSourceChange(tx, changes, false); err != nil {
			return err
		}
		updated, err = getPostgresV1Transaction(tx, input.UserId, input.TransactionId)
		return err
	})
	return updated, previousDate, err
}

func (ts *TransactionStore) DeleteV1Transaction(userId string, transactionId string) error {
	return ts.DeleteV1TransactionVersion(userId, transactionId, nil)
}
func (ts *TransactionStore) DeleteV1TransactionVersion(userId, transactionId string, version *int64) error {
	return householdTransaction(ts.db, func(tx *gorm.DB) error {
		changes, err := beginHouseholdSourceChange(tx, userId, transactionId)
		if err != nil {
			return err
		}
		current, err := getPostgresV1Transaction(tx, userId, transactionId)
		if err != nil {
			return err
		}
		if version != nil && *version != current.Version {
			return household.Conflict
		}
		if err = tx.Table("transaction").Where("transaction_id = ? AND user_no = ? AND deleted_at IS NULL", transactionId, userId).Updates(map[string]any{"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"), "updated_at": gorm.Expr("CURRENT_TIMESTAMP"), "version": gorm.Expr("version + 1")}).Error; err != nil {
			return err
		}
		return finishHouseholdSourceChange(tx, changes, true)
	})
}

func (ts *TransactionStore) GetV1AnalyticsTransactions(userId string, startDate string, endDate string) ([]model.V1AnalyticsTransaction, error) {
	result := make([]model.V1AnalyticsTransaction, 0)
	err := ts.db.Table("transaction t").Where("t.deleted_at IS NULL").
		Select(
			"CAST(t.transaction_id AS TEXT) AS transaction_id",
			"TO_CHAR(t.transaction_date, 'YYYY-MM-DD') AS transaction_date",
			v1TransactionTimeSelect,
			"t.transaction_name",
			"t.transaction_amount AS signed_amount",
			"CAST(t.category_id AS TEXT) AS category_id",
			"c.category_name",
			"CAST(t.sub_category_id AS TEXT) AS sub_category_id",
			"sc.sub_category_name",
			"t.fixed_flg",
			"CAST(t.payment_id AS TEXT) AS payment_id",
			"pr.payment_name",
			"CAST(pr.payment_type_id AS TEXT) AS payment_type_id",
			"pt.payment_type_name",
			"pt.is_payment_due_later",
		).
		Joins("INNER JOIN category c ON c.category_id = t.category_id").
		Joins("INNER JOIN sub_category sc ON sc.sub_category_id = t.sub_category_id").
		Joins("LEFT JOIN payment_resource pr ON pr.payment_id = t.payment_id").
		Joins("LEFT JOIN payment_type pt ON pt.payment_type_id = pr.payment_type_id").
		Where("t.user_no = ?", userId).
		Where("t.transaction_date BETWEEN ? AND ?", startDate, endDate).
		Order("t.transaction_date DESC, t.transaction_time DESC NULLS LAST, t.transaction_id DESC").
		Scan(&result).Error
	return result, err
}

func validatePostgresV1BaseRelations(db *gorm.DB, input *model.V1TransactionWrite) error {
	var categoryCount int64
	if err := db.Table("category").
		Where("category_id = ?", input.CategoryId).
		Count(&categoryCount).Error; err != nil {
		return err
	}
	if categoryCount != 1 {
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

func validatePostgresV1SubCategory(db *gorm.DB, input *model.V1TransactionWrite) error {
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
	return nil
}

type v1SubCategoryCandidate struct {
	SubCategoryId string `gorm:"column:sub_category_id"`
	UserId        string `gorm:"column:user_no"`
	Enable        bool   `gorm:"column:enable"`
}

func resolveV1WriteSubCategory(db *gorm.DB, input *model.V1TransactionWrite) (string, error) {
	var candidate v1SubCategoryCandidate
	query := db.Table("sub_category sc").
		Select("CAST(sc.sub_category_id AS TEXT) AS sub_category_id, CAST(sc.user_no AS TEXT) AS user_no, NOT EXISTS (SELECT 1 FROM hidden_sub_category hsc WHERE hsc.sub_category_id = sc.sub_category_id AND hsc.user_no = ?) AS enable", input.UserId).
		Where("sc.category_id = ?", input.CategoryId).
		Where("sc.sub_category_name = ?", input.SubCategoryName).
		Where("sc.user_no IN ?", []string{"1", input.UserId}).
		Order("enable DESC").
		Order("CASE WHEN sc.user_no = 1 THEN 0 ELSE 1 END").
		Order("sc.sub_category_id")
	err := query.Take(&candidate).Error
	if err == nil {
		if !candidate.Enable {
			if err := db.Table("hidden_sub_category").
				Where("user_no = ?", input.UserId).
				Where("sub_category_id = ?", candidate.SubCategoryId).
				Delete(&model.EditSubCategoryModel{}).Error; err != nil {
				return "", fmt.Errorf("%w: %w", subcategorydomain.ErrResolveFailed, err)
			}
		}
		return candidate.SubCategoryId, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("%w: %w", subcategorydomain.ErrResolveFailed, err)
	}

	created := model.SubCategoryModel{
		UserNo:          input.UserId,
		CategoryId:      input.CategoryId,
		SubCategoryName: input.SubCategoryName,
	}
	if err := db.Table("sub_category").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_no"}, {Name: "category_id"}, {Name: "sub_category_name"}},
		DoNothing: true,
	}).Create(&created).Error; err != nil {
		return "", fmt.Errorf("%w: %w", subcategorydomain.ErrResolveFailed, err)
	}
	if err := db.Table("sub_category").
		Where("user_no = ?", input.UserId).
		Where("category_id = ?", input.CategoryId).
		Where("sub_category_name = ?", input.SubCategoryName).
		Take(&created).Error; err != nil {
		return "", fmt.Errorf("%w: %w", subcategorydomain.ErrResolveFailed, err)
	}
	return strconv.FormatInt(created.SubCategoryId, 10), nil
}

func stringID(id uint64) string {
	return strconv.FormatUint(id, 10)
}
