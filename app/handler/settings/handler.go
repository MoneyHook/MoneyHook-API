package settings

import settingsdomain "MoneyHook/MoneyHook-API/settings"

type Handler struct {
	settingsStore settingsdomain.Store
}

func New(settingsStore settingsdomain.Store) *Handler {
	return &Handler{settingsStore: settingsStore}
}
