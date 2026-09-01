package budget

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

type fakeBudgetStore struct {
	budgets map[string]model.Budget
}

func (s *fakeBudgetStore) GetBudget(_ string, month string) (*model.Budget, error) {
	var selected *model.Budget
	for _, budget := range s.budgets {
		if budget.EffectiveFrom <= month && (selected == nil || budget.EffectiveFrom > selected.EffectiveFrom) {
			candidate := budget
			selected = &candidate
		}
	}
	if selected == nil {
		return nil, nil
	}
	return selected, nil
}

func (s *fakeBudgetStore) UpsertBudget(input *model.Budget) error {
	if s.budgets == nil {
		s.budgets = make(map[string]model.Budget)
	}
	s.budgets[input.EffectiveFrom] = *input
	return nil
}

func TestValidateV1BudgetRequest(t *testing.T) {
	valid := v1BudgetRequest{MonthlyBudgetAmount: 300000, EffectiveFrom: "2026-09-01"}
	if fieldErrors := validateV1BudgetRequest(valid); len(fieldErrors) != 0 {
		t.Fatalf("valid request returned errors: %v", fieldErrors)
	}

	invalid := v1BudgetRequest{MonthlyBudgetAmount: 0, EffectiveFrom: "2026-09-02"}
	fieldErrors := validateV1BudgetRequest(invalid)
	for _, field := range []string{"monthly_budget_amount", "effective_from"} {
		if _, exists := fieldErrors[field]; !exists {
			t.Errorf("missing validation error for %s: %v", field, fieldErrors)
		}
	}
}

func TestIsMonthStart(t *testing.T) {
	for _, value := range []string{"2026-01-01", "2026-02-01", "2026-12-01"} {
		if !isMonthStart(value) {
			t.Errorf("%q should be a valid start month", value)
		}
	}
	for _, value := range []string{"", "2026-02-02", "2026-02-30", "2026/02/01", "2026-2-01"} {
		if isMonthStart(value) {
			t.Errorf("%q should be an invalid start month", value)
		}
	}
}

func TestNewV1BudgetResponse(t *testing.T) {
	unset := newV1BudgetResponse(nil)
	if unset.MonthlyBudgetAmount != nil || unset.EffectiveFrom != nil {
		t.Fatalf("unset budget response = %+v, want nullable fields", unset)
	}
}

func TestGetV1BudgetReturnsNullWhenUnset(t *testing.T) {
	h := New(&fakeBudgetStore{})
	e := echo.New()
	recorder := httptest.NewRecorder()
	context := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/budget?month=2026-09-01", nil), recorder)
	context.Set(router.ContextKeyUserNo, "42")

	if err := h.GetV1Budget(context); err != nil {
		t.Fatalf("GetV1Budget() error = %v", err)
	}
	var response v1BudgetResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.MonthlyBudgetAmount != nil || response.EffectiveFrom != nil {
		t.Fatalf("unset budget response = %+v", response)
	}
}

func TestGetV1BudgetRejectsInvalidMonth(t *testing.T) {
	h := New(&fakeBudgetStore{})
	e := echo.New()
	recorder := httptest.NewRecorder()
	context := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/budget?month=2026-09-02", nil), recorder)
	context.Set(router.ContextKeyUserNo, "42")

	if err := h.GetV1Budget(context); err != nil {
		t.Fatalf("GetV1Budget() error = %v", err)
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("invalid month status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestPutV1BudgetUsesAuthenticatedUserAndPreservesHistory(t *testing.T) {
	store := &fakeBudgetStore{}
	h := New(store)
	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/api/v1/budget", strings.NewReader(`{"monthly_budget_amount":300000,"effective_from":"2026-07-01"}`))
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	context := e.NewContext(request, recorder)
	context.Set(router.ContextKeyUserNo, "42")

	if err := h.PutV1Budget(context); err != nil {
		t.Fatalf("PutV1Budget() error = %v", err)
	}
	storedJuly := store.budgets["2026-07-01"]
	if storedJuly.UserNo != "42" || storedJuly.MonthlyBudgetAmount != 300000 || storedJuly.EffectiveFrom != "2026-07-01" {
		t.Fatalf("stored July budget = %+v", storedJuly)
	}

	secondRecorder := httptest.NewRecorder()
	secondContext := e.NewContext(httptest.NewRequest(http.MethodPut, "/api/v1/budget", strings.NewReader(`{"monthly_budget_amount":350000,"effective_from":"2026-08-01"}`)), secondRecorder)
	secondContext.Set(router.ContextKeyUserNo, "42")
	if err := h.PutV1Budget(secondContext); err != nil {
		t.Fatalf("second PutV1Budget() error = %v", err)
	}
	if store.budgets["2026-07-01"].MonthlyBudgetAmount != 300000 || store.budgets["2026-08-01"].MonthlyBudgetAmount != 350000 {
		t.Fatalf("budget history = %+v", store.budgets)
	}

	getJulyRecorder := httptest.NewRecorder()
	getJulyContext := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/budget?month=2026-07-01", nil), getJulyRecorder)
	getJulyContext.Set(router.ContextKeyUserNo, "42")
	if err := h.GetV1Budget(getJulyContext); err != nil {
		t.Fatalf("GetV1Budget() for July error = %v", err)
	}
	var julyResponse v1BudgetResponse
	if err := json.Unmarshal(getJulyRecorder.Body.Bytes(), &julyResponse); err != nil {
		t.Fatalf("decode July response: %v", err)
	}
	if julyResponse.MonthlyBudgetAmount == nil || *julyResponse.MonthlyBudgetAmount != 300000 || julyResponse.EffectiveFrom == nil || *julyResponse.EffectiveFrom != "2026-07-01" {
		t.Fatalf("July budget response = %+v", julyResponse)
	}

	getAugustRecorder := httptest.NewRecorder()
	getAugustContext := e.NewContext(httptest.NewRequest(http.MethodGet, "/api/v1/budget?month=2026-08-01", nil), getAugustRecorder)
	getAugustContext.Set(router.ContextKeyUserNo, "42")
	if err := h.GetV1Budget(getAugustContext); err != nil {
		t.Fatalf("GetV1Budget() for August error = %v", err)
	}
	var augustResponse v1BudgetResponse
	if err := json.Unmarshal(getAugustRecorder.Body.Bytes(), &augustResponse); err != nil {
		t.Fatalf("decode August response: %v", err)
	}
	if augustResponse.MonthlyBudgetAmount == nil || *augustResponse.MonthlyBudgetAmount != 350000 || augustResponse.EffectiveFrom == nil || *augustResponse.EffectiveFrom != "2026-08-01" {
		t.Fatalf("August budget response = %+v", augustResponse)
	}

	unauthenticatedRecorder := httptest.NewRecorder()
	if err := h.PutV1Budget(e.NewContext(httptest.NewRequest(http.MethodPut, "/api/v1/budget", nil), unauthenticatedRecorder)); err != nil {
		t.Fatalf("unauthenticated PutV1Budget() error = %v", err)
	}
	if unauthenticatedRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want %d", unauthenticatedRecorder.Code, http.StatusUnauthorized)
	}
}
