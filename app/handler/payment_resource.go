package handler

import (
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetPaymentResourceList(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	result := h.paymentResourceStore.GetPaymentResourceList(userId)

	for _, item := range *result {
		fmt.Println(item)
	}

	fmt.Println(result)

	result_list := getPaymentResourceListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}
