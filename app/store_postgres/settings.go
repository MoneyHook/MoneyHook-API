package store_postgres

import (
	household "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/model"

	"gorm.io/gorm"
)

type SettingsStore struct {
	db *gorm.DB
}

func NewSettingsStore(db *gorm.DB) *SettingsStore {
	return &SettingsStore{db: db}
}

type postgresUserSettingsRecord struct {
	DefaultTransactionScope string `gorm:"column:default_transaction_scope"`
	AccentColor             string `gorm:"column:accent_color"`
	ThemeMode               string `gorm:"column:theme_mode"`
	ChartPalette            string `gorm:"column:chart_palette"`
}

func (ss *SettingsStore) GetSettings(userNo string) (*model.UserSettings, error) {
	var record postgresUserSettingsRecord
	if err := ss.db.Table("users").
		Select("accent_color", "theme_mode", "chart_palette", "default_transaction_scope").
		Where("user_no = ?", userNo).
		Take(&record).Error; err != nil {
		return nil, err
	}
	return &model.UserSettings{DefaultTransactionScope: record.DefaultTransactionScope, AccentColor: record.AccentColor, ThemeMode: record.ThemeMode, ChartPalette: record.ChartPalette}, nil
}

func (ss *SettingsStore) UpdateSettings(userNo string, update *model.UserSettingsUpdate) (*model.UserSettings, error) {
	err := householdTransaction(ss.db, func(tx *gorm.DB) error {
		if err := lockHouseholdUser(tx, userNo); err != nil {
			return err
		}
		values := map[string]any{}
		if update.AccentColor != nil {
			values["accent_color"] = *update.AccentColor
		}
		if update.ThemeMode != nil {
			values["theme_mode"] = *update.ThemeMode
		}
		if update.ChartPalette != nil {
			values["chart_palette"] = *update.ChartPalette
		}
		if update.DefaultTransactionScope != nil {
			if *update.DefaultTransactionScope == "household" {
				var n int64
				if err := tx.Table("household_member").Where("user_no = ? AND state='active'", userNo).Count(&n).Error; err != nil {
					return err
				}
				if n != 1 {
					return household.Invalid
				}
			}
			values["default_transaction_scope"] = *update.DefaultTransactionScope
		}
		if len(values) == 0 {
			return nil
		}
		return tx.Table("users").Where("user_no = ?", userNo).Updates(values).Error
	})
	if err != nil {
		return nil, err
	}
	return ss.GetSettings(userNo)
}
