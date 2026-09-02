package model

type UserSettings struct {
	AccentColor  string
	ThemeMode    string
	ChartPalette string
}

type UserSettingsUpdate struct {
	AccentColor  *string
	ThemeMode    *string
	ChartPalette *string
}
