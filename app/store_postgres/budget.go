package store_postgres

import (
	"MoneyHook/MoneyHook-API/model"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BudgetStore struct {
	db *gorm.DB
}

func NewBudgetStore(db *gorm.DB) *BudgetStore {
	return &BudgetStore{db: db}
}

type postgresBudgetRecord struct {
	UserNo              string `gorm:"column:user_no"`
	MonthlyBudgetAmount int64  `gorm:"column:monthly_budget_amount"`
	EffectiveFrom       string `gorm:"column:effective_from"`
}

func (bs *BudgetStore) GetBudget(userNo string, month string) (*model.Budget, error) {
	var record postgresBudgetRecord
	err := bs.db.Table("budget").
		Select("CAST(user_no AS TEXT) AS user_no", "monthly_budget_amount", "TO_CHAR(effective_from, 'YYYY-MM-DD') AS effective_from").
		Where("user_no = ?", userNo).
		Where("effective_from <= ?", month).
		Order("effective_from DESC").
		Take(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &model.Budget{
		UserNo:              record.UserNo,
		MonthlyBudgetAmount: record.MonthlyBudgetAmount,
		EffectiveFrom:       record.EffectiveFrom,
	}, nil
}

func (bs *BudgetStore) UpsertBudget(input *model.Budget) error {
	return bs.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_no"}, {Name: "effective_from"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"monthly_budget_amount",
		}),
	}).Table("budget").Create(map[string]any{
		"user_no":               input.UserNo,
		"monthly_budget_amount": input.MonthlyBudgetAmount,
		"effective_from":        input.EffectiveFrom,
	}).Error
}
