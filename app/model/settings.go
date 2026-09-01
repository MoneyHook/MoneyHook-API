package model

type UserSettings struct {
	AccentColor string
	ThemeMode   string
}

type UserSettingsUpdate struct {
	AccentColor *string
	ThemeMode   *string
}
