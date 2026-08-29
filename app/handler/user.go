package handler

import (
	"MoneyHook/MoneyHook-API/router"
	"errors"

	"github.com/labstack/echo/v4"
)

var errAuthenticatedUserMissing = errors.New("authenticated user is missing from request context")

func (h *Handler) GetUserId(c echo.Context) (string, error) {
	userNo := router.GetUserNo(c)
	if userNo == "" {
		return "", errAuthenticatedUserMissing
	}
	return userNo, nil
}

func (h *Handler) GetV1UserId(c echo.Context) (string, error) {
	return h.GetUserId(c)
}
