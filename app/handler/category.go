package handler

import (
	"MoneyHook/MoneyHook-API/handler/response"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetCategoryList(c echo.Context) error {
	result := h.categoryStore.GetCategoryList()

	result_list := response.GetCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) GetCategoryWithSubCategoryList(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	result := h.categoryStore.GetCategoryWithSubCategoryList(userId)

	result_list := response.GetCategoryWithSubCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}
