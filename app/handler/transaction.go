package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) getTimelineData(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetTimelineData(userId, month)

	result_list := getTimelineListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}
