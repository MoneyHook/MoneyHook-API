package httpx

import (
	"MoneyHook/MoneyHook-API/router"
	"errors"

	"github.com/labstack/echo/v4"
)

var errAuthenticatedUserMissing = errors.New("authenticated user is missing from request context")

func UserID(c echo.Context) (string, error) {
	userNo := router.GetUserNo(c)
	if userNo == "" {
		return "", errAuthenticatedUserMissing
	}
	return userNo, nil
}
