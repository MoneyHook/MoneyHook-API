package handler

import (
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetPaymentResourceList(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	result := h.paymentResourceStore.GetPaymentResourceList(userId)

	result_list := getPaymentResourceListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) AddPaymentResource(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var addPaymentResource model.AddPaymentResource

	addPaymentResource.UserNo = userId

	req := &AddPaymentRequest{}
	if err := req.bind(c, &addPaymentResource); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
	}

	err = h.paymentResourceStore.AddPaymentResource(&addPaymentResource)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) DeletePaymentResource(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	paymentId, err := strconv.Atoi(c.Param("paymentId"))
	if err != nil {
		return c.JSON(http.StatusOK, "hej")
	}

	deletePaymentResource := model.DeletePaymentResource{UserNo: userId, PaymentId: paymentId}

	err = h.paymentResourceStore.DeletePaymentResource(&deletePaymentResource)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}
