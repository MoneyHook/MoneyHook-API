package budget

import (
	"MoneyHook/MoneyHook-API/handler/internal/httpx"
	"MoneyHook/MoneyHook-API/model"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type v1BudgetRequest struct {
	MonthlyBudgetAmount int64  `json:"monthly_budget_amount"`
	EffectiveFrom       string `json:"effective_from"`
}

type v1BudgetResponse struct {
	MonthlyBudgetAmount *int64  `json:"monthly_budget_amount"`
	EffectiveFrom       *string `json:"effective_from"`
}

func (h *Handler) GetV1Budget(c echo.Context) error {
	userNo, err := httpx.UserID(c)
	if err != nil {
		return httpx.RespondV1Unauthorized(c)
	}
	month := c.QueryParam("month")
	if !isMonthStart(month) {
		return httpx.RespondV1Error(c, http.StatusBadRequest, "INVALID_QUERY", "クエリパラメータが不正です", map[string]string{
			"month": "YYYY-MM-01形式の実在する月初日を指定してください",
		})
	}
	result, err := h.budgetStore.GetBudget(userNo, month)
	if err != nil {
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "予算設定の取得に失敗しました", nil)
	}
	return c.JSON(http.StatusOK, newV1BudgetResponse(result))
}

func (h *Handler) PutV1Budget(c echo.Context) error {
	userNo, err := httpx.UserID(c)
	if err != nil {
		return httpx.RespondV1Unauthorized(c)
	}
	var request v1BudgetRequest
	if err := httpx.DecodeV1JSON(c, &request); err != nil {
		return httpx.RespondV1Error(c, http.StatusBadRequest, "INVALID_JSON", "JSON形式が不正です", nil)
	}
	fieldErrors := validateV1BudgetRequest(request)
	if len(fieldErrors) > 0 {
		return httpx.RespondV1Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "入力内容を確認してください", fieldErrors)
	}

	input := model.Budget{
		UserNo:              userNo,
		MonthlyBudgetAmount: request.MonthlyBudgetAmount,
		EffectiveFrom:       request.EffectiveFrom,
	}
	if err := h.budgetStore.UpsertBudget(&input); err != nil {
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "予算設定の保存に失敗しました", nil)
	}
	return c.JSON(http.StatusOK, newV1BudgetResponse(&input))
}

func newV1BudgetResponse(budget *model.Budget) v1BudgetResponse {
	if budget == nil {
		return v1BudgetResponse{}
	}
	amount := budget.MonthlyBudgetAmount
	effectiveFrom := budget.EffectiveFrom
	return v1BudgetResponse{
		MonthlyBudgetAmount: &amount,
		EffectiveFrom:       &effectiveFrom,
	}
}

func validateV1BudgetRequest(request v1BudgetRequest) map[string]string {
	fieldErrors := map[string]string{}
	if request.MonthlyBudgetAmount < 1 {
		fieldErrors["monthly_budget_amount"] = "1以上の金額を指定してください"
	}
	if !isMonthStart(request.EffectiveFrom) {
		fieldErrors["effective_from"] = "YYYY-MM-01形式の実在する月初日を指定してください"
	}
	return fieldErrors
}

func isMonthStart(value string) bool {
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Day() == 1 && parsed.Format("2006-01-02") == value
}
