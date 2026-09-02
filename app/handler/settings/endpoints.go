package settings

import (
	"MoneyHook/MoneyHook-API/handler/internal/httpx"
	"MoneyHook/MoneyHook-API/model"
	"net/http"

	"github.com/labstack/echo/v4"
)

var validAccentColors = map[string]struct{}{
	"blue":   {},
	"green":  {},
	"violet": {},
	"rose":   {},
	"black":  {},
}

var validThemeModes = map[string]struct{}{
	"light":  {},
	"dark":   {},
	"system": {},
}

var validChartPalettes = map[string]struct{}{
	"default":    {},
	"colorful":   {},
	"monochrome": {},
}

type v1SettingsPatchRequest struct {
	AccentColor  *string `json:"accent_color"`
	ThemeMode    *string `json:"theme_mode"`
	ChartPalette *string `json:"chart_palette"`
}

type v1SettingsResponse struct {
	AccentColor  string `json:"accent_color"`
	ThemeMode    string `json:"theme_mode"`
	ChartPalette string `json:"chart_palette"`
}

func (h *Handler) GetV1Settings(c echo.Context) error {
	userNo, err := httpx.UserID(c)
	if err != nil {
		return httpx.RespondV1Unauthorized(c)
	}
	result, err := h.settingsStore.GetSettings(userNo)
	if err != nil {
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "表示設定の取得に失敗しました", nil)
	}
	return c.JSON(http.StatusOK, newV1SettingsResponse(result))
}

func (h *Handler) PatchV1Settings(c echo.Context) error {
	userNo, err := httpx.UserID(c)
	if err != nil {
		return httpx.RespondV1Unauthorized(c)
	}
	var request v1SettingsPatchRequest
	if err := httpx.DecodeV1JSON(c, &request); err != nil {
		return httpx.RespondV1Error(c, http.StatusBadRequest, "INVALID_JSON", "JSON形式が不正です", nil)
	}
	fieldErrors := validateV1SettingsPatch(request)
	if len(fieldErrors) > 0 {
		return httpx.RespondV1Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "入力内容を確認してください", fieldErrors)
	}

	result, err := h.settingsStore.UpdateSettings(userNo, &model.UserSettingsUpdate{
		AccentColor:  request.AccentColor,
		ThemeMode:    request.ThemeMode,
		ChartPalette: request.ChartPalette,
	})
	if err != nil {
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "表示設定の保存に失敗しました", nil)
	}
	return c.JSON(http.StatusOK, newV1SettingsResponse(result))
}

func newV1SettingsResponse(settings *model.UserSettings) v1SettingsResponse {
	return v1SettingsResponse{
		AccentColor:  settings.AccentColor,
		ThemeMode:    settings.ThemeMode,
		ChartPalette: settings.ChartPalette,
	}
}

func validateV1SettingsPatch(request v1SettingsPatchRequest) map[string]string {
	fieldErrors := map[string]string{}
	if request.AccentColor == nil && request.ThemeMode == nil && request.ChartPalette == nil {
		fieldErrors["settings"] = "少なくとも1つの設定項目を指定してください"
	}
	if request.AccentColor != nil {
		if _, ok := validAccentColors[*request.AccentColor]; !ok {
			fieldErrors["accent_color"] = "blue、green、violet、rose、blackのいずれかを指定してください"
		}
	}
	if request.ThemeMode != nil {
		if _, ok := validThemeModes[*request.ThemeMode]; !ok {
			fieldErrors["theme_mode"] = "light、dark、systemのいずれかを指定してください"
		}
	}
	if request.ChartPalette != nil {
		if _, ok := validChartPalettes[*request.ChartPalette]; !ok {
			fieldErrors["chart_palette"] = "default、colorful、monochromeのいずれかを指定してください"
		}
	}
	return fieldErrors
}
