package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetCategoryList(c echo.Context) error {
	result := h.categoryStore.GetCategoryList()

	result_list := getCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) GetCategoryWithSubCategoryList(c echo.Context) error {
	userId := h.GetUserId(c)
	result := h.categoryStore.GetCategoryWithSubCategoryList(userId)

	result_list := getCategoryWithSubCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}
