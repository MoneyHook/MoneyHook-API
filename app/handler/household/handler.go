package household

import (
	"MoneyHook/MoneyHook-API/handler/internal/httpx"
	h "MoneyHook/MoneyHook-API/household"
	t "MoneyHook/MoneyHook-API/transaction"
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"net"
	"net/http"
	"strconv"
)

type Handler struct{ store h.Store }

func New(store h.Store) *Handler { return &Handler{store: store} }

type call func(context.Context, string, string) (any, error)

func (v *Handler) run(c echo.Context, status int, fn call) error {
	u, e := httpx.UserID(c)
	if e != nil {
		return httpx.RespondV1Unauthorized(c)
	}
	for _, name := range []string{"householdId", "memberId", "invitationId", "entryId", "transactionId", "referenceId"} {
		if id := c.Param(name); id != "" && !h.ID(id) {
			return httpx.RespondV1Error(c, 400, "INVALID_PATH_PARAMETER", "IDが不正です", nil)
		}
	}
	out, e := fn(c.Request().Context(), u, c.Param("householdId"))
	if e != nil {
		return respond(c, e)
	}
	if status == 204 {
		return c.NoContent(204)
	}
	return c.JSON(status, out)
}
func respond(c echo.Context, e error) error {
	if errors.Is(e, t.ErrNotFound) {
		return httpx.RespondV1Error(c, 404, "NOT_FOUND", "取引が見つかりません", nil)
	}
	if errors.Is(e, t.ErrInvalidRelation) {
		return httpx.RespondV1Error(c, 422, "VALIDATION_ERROR", "利用可能なカテゴリ、サブカテゴリ、支払い方法を選択してください", nil)
	}
	var err *h.Error
	if !errors.As(e, &err) {
		return httpx.RespondV1Error(c, 500, "INTERNAL_ERROR", "処理に失敗しました", nil)
	}
	status := 409
	switch err.Code {
	case "NOT_FOUND":
		status = 404
	case "FORBIDDEN":
		status = 403
	case "VALIDATION_ERROR", "INVITATION_UNAVAILABLE":
		status = 422
	case "RATE_LIMITED":
		status = 429
		c.Response().Header().Set("Retry-After", "900")
	case "INTERNAL_ERROR":
		status = 500
	case "INVITATION_NOT_CONFIGURED":
		status = 503
	}
	return httpx.RespondV1Error(c, status, err.Code, err.Message, nil)
}
func read[T any](c echo.Context) (T, error) {
	var v T
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, 65536)
	if e := httpx.DecodeV1JSON(c, &v); e != nil {
		return v, httpx.RespondV1Error(c, 400, "INVALID_JSON", "入力形式が不正です", nil)
	}
	return v, nil
}

// Socket peer only: forwarded headers are not trusted without deployment configuration.
func peer(c echo.Context) string {
	host, _, e := net.SplitHostPort(c.Request().RemoteAddr)
	if e != nil {
		return c.Request().RemoteAddr
	}
	return host
}
func version(c echo.Context) int64 {
	v, _ := strconv.ParseInt(c.QueryParam("expected_version"), 10, 64)
	return v
}
func key(c echo.Context) string { return c.Request().Header.Get("Idempotency-Key") }
