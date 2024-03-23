package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (h *Handler) getTimelineData(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetTimelineData(userId, month)

	result_list := getTimelineListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getMonthlySpendingData(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlySpendingData(userId, month)

	result_list := getmonthlySpendingDataResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getTransaction(c echo.Context) error {
	userId := getUserId(c)
	transactionId, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		return c.JSON(http.StatusOK, "hej")
	}
	result := h.transactionStore.GetTransactionData(userId, transactionId)

	result_list := getTransactionResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getMonthlyFixedIncome(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyFixedData(userId, month, false)

	result_list := getMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyFixedSpending(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyFixedData(userId, month, true)

	result_list := getMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getHome(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetHome(userId, month)

	result_list := getHomeResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyVariableData(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyVariableData(userId, month)

	result_list := getMonthlyVariableResponse(result)

	return c.JSON(http.StatusOK, result_list)
}
