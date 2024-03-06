package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetSubCategoryList(c echo.Context) error {
	userId := getUserId(c)
	categoryId, err := strconv.Atoi(c.Param("categoryId"))
	if err != nil {
		return c.JSON(http.StatusOK, "hej")
	}

	result := h.subCategoryStore.GetSubCategoryList(userId, categoryId)

	result_list := getSubCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}
