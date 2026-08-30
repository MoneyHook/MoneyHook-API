package settings

import (
	"MoneyHook/MoneyHook-API/model"
	"MoneyHook/MoneyHook-API/router"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type fakeSettingsStore struct {
	settings model.UserSettings
}

func (s *fakeSettingsStore) GetSettings(string) (*model.UserSettings, error) {
	result := s.settings
	return &result, nil
}

func (s *fakeSettingsStore) UpdateSettings(_ string, update *model.UserSettingsUpdate) (*model.UserSettings, error) {
	if update.AccentColor != nil {
		s.settings.AccentColor = *update.AccentColor
	}
	if update.ThemeMode != nil {
		s.settings.ThemeMode = *update.ThemeMode
	}
	return s.GetSettings("")
}

func TestValidateV1SettingsPatch(t *testing.T) {
	accent := "violet"
	mode := "dark"
	if fieldErrors := validateV1SettingsPatch(v1SettingsPatchRequest{AccentColor: &accent, ThemeMode: &mode}); len(fieldErrors) != 0 {
		t.Fatalf("valid patch returned errors: %v", fieldErrors)
	}

	invalidAccent := "orange"
	invalidMode := "auto"
	fieldErrors := validateV1SettingsPatch(v1SettingsPatchRequest{AccentColor: &invalidAccent, ThemeMode: &invalidMode})
	for _, field := range []string{"accent_color", "theme_mode"} {
		if _, exists := fieldErrors[field]; !exists {
			t.Errorf("missing validation error for %s: %v", field, fieldErrors)
		}
	}

	if fieldErrors := validateV1SettingsPatch(v1SettingsPatchRequest{}); len(fieldErrors) == 0 {
		t.Fatal("empty patch should return a validation error")
	}
}

func TestPatchV1SettingsPreservesOmittedValue(t *testing.T) {
	store := &fakeSettingsStore{settings: model.UserSettings{AccentColor: "blue", ThemeMode: "system"}}
	h := New(store)
	e := echo.New()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", strings.NewReader(`{"accent_color":"rose"}`))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)
	context.Set(router.ContextKeyUserNo, "42")

	if err := h.PatchV1Settings(context); err != nil {
		t.Fatalf("PatchV1Settings() error = %v", err)
	}
	if context.Response().Status != http.StatusOK {
		t.Fatalf("status = %d, want %d", context.Response().Status, http.StatusOK)
	}
	if store.settings.AccentColor != "rose" || store.settings.ThemeMode != "system" {
		t.Fatalf("settings after partial update = %+v", store.settings)
	}

	var response v1SettingsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.AccentColor != "rose" || response.ThemeMode != "system" {
		t.Errorf("response = %+v", response)
	}
}
