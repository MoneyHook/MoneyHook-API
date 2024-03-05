package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func getUserId(c echo.Context) int {
	userId, error := strconv.Atoi(c.Request().Header["Authorization"][0])
	if error != nil {
		return 1
	}
	return userId
}

func (h *Handler) GetCategoryList(c echo.Context) error {
	result := h.categoryStore.GetCategoryList()

	result_list := getCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) GetCategoryWithSubCategoryList(c echo.Context) error {
	userId := getUserId(c)
	result := h.categoryStore.GetCategoryWithSubCategoryList(userId)

	result_list := getCategoryWithSubCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}
