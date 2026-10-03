package model

type UserSettings struct {
	DefaultTransactionScope string
	AccentColor             string
	ThemeMode               string
	ChartPalette            string
}

type UserSettingsUpdate struct {
	DefaultTransactionScope *string
	AccentColor             *string
	ThemeMode               *string
	ChartPalette            *string
}
