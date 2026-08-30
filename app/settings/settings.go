package settings

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetSettings(userNo string) (*model.UserSettings, error)
	UpdateSettings(userNo string, update *model.UserSettingsUpdate) (*model.UserSettings, error)
}
