package transaction

import (
	"MoneyHook/MoneyHook-API/handler/internal/httpx"
	household "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/model"
	"errors"
	"github.com/labstack/echo/v4"
)

func (h *Handler) ListV1Transactions(c echo.Context) error {
	u, e := httpx.UserID(c)
	if e != nil {
		return httpx.RespondV1Unauthorized(c)
	}
	s, ok := h.transactionStore.(interface {
		ListV1Transactions(string, string, string, string) ([]model.V1Transaction, *string, error)
	})
	if !ok {
		return httpx.RespondV1Error(c, 500, "INTERNAL_ERROR", "取引を取得できません", nil)
	}
	rows, next, e := s.ListV1Transactions(u, c.QueryParam("month"), c.QueryParam("sharing"), c.QueryParam("cursor"))
	if e != nil {
		if errors.Is(e, household.Invalid) {
			return httpx.RespondV1Error(c, 422, "VALIDATION_ERROR", e.Error(), nil)
		}
		return h.respondV1TransactionStoreError(c, e)
	}
	out := []v1TransactionResource{}
	for i := range rows {
		out = append(out, newV1TransactionResource(&rows[i]))
	}
	return c.JSON(200, struct {
		Transactions []v1TransactionResource `json:"transactions"`
		NextCursor   *string                 `json:"next_cursor"`
	}{out, next})
}
