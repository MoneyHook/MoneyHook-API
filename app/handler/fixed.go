package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) getFixed(c echo.Context) error {
	userId := getUserId(c)

	result := h.fixedStore.GetFixedData(userId)

	result_list := GetFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}
