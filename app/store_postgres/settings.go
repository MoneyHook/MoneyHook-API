package store_postgres

import (
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
	AccentColor string `gorm:"column:accent_color"`
	ThemeMode   string `gorm:"column:theme_mode"`
}

func (ss *SettingsStore) GetSettings(userNo string) (*model.UserSettings, error) {
	var record postgresUserSettingsRecord
	if err := ss.db.Table("users").
		Select("accent_color", "theme_mode").
		Where("user_no = ?", userNo).
		Take(&record).Error; err != nil {
		return nil, err
	}
	return &model.UserSettings{AccentColor: record.AccentColor, ThemeMode: record.ThemeMode}, nil
}

func (ss *SettingsStore) UpdateSettings(userNo string, update *model.UserSettingsUpdate) (*model.UserSettings, error) {
	values := map[string]any{}
	if update.AccentColor != nil {
		values["accent_color"] = *update.AccentColor
	}
	if update.ThemeMode != nil {
		values["theme_mode"] = *update.ThemeMode
	}
	if len(values) > 0 {
		if err := ss.db.Table("users").Where("user_no = ?", userNo).Updates(values).Error; err != nil {
			return nil, err
		}
	}
	return ss.GetSettings(userNo)
}
