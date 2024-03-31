package handler

import (
	"MoneyHook/MoneyHook-API/model"
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

func (h *Handler) EditSubCategory(c echo.Context) error {
	userId := getUserId(c)
	var editSubCategory model.EditSubCategoryModel

	editSubCategory.UserId = userId

	req := &editSubCategoryRequest{}
	if err := req.bind(c, &editSubCategory); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if editSubCategory.IsEnable {
		// err := h.subCategoryStore.ExposeSubCategory(&editSubCategory)
		h.subCategoryStore.ExposeSubCategory(&editSubCategory)
	} else {
		// err := h.subCategoryStore.HideSubCategory(&editSubCategory)
		h.subCategoryStore.HideSubCategory(&editSubCategory)
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}
