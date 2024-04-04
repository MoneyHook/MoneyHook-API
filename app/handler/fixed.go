package handler

import (
	"MoneyHook/MoneyHook-API/model"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) getFixed(c echo.Context) error {
	userId := getUserId(c)

	result := h.fixedStore.GetFixedData(userId)

	result_list := GetFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getDeletedFixed(c echo.Context) error {
	userId := getUserId(c)

	result := h.fixedStore.GetFixedDeletedData(userId)

	result_list := GetFixedDeletedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) addFixed(c echo.Context) error {
	userId := getUserId(c)
	var addFixed model.AddFixed

	addFixed.UserId = userId

	req := &addFixedRequest{}
	if err := req.bind(c, &addFixed); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if addFixed.SubCategoryName != "" {
		subCategory := model.SubCategoryModel{
			UserNo:          addFixed.UserId,
			CategoryId:      addFixed.CategoryId,
			SubCategoryName: addFixed.SubCategoryName,
		}
		// TODO Createの前に、同じユーザー、同じカテゴリIDに紐づくサブカテゴリ名が存在するか確認
		h.subCategoryStore.CreateSubCategory(&subCategory)
		addFixed.SubCategoryId = subCategory.SubCategoryId
	}

	// err := h.FixedStore.AddFixed(&addFixed)
	h.fixedStore.AddFixed(&addFixed)

	return c.JSON(http.StatusOK, "ok")
}

func (h *Handler) editFixed(c echo.Context) error {
	userId := getUserId(c)
	var editFixed model.EditFixed

	editFixed.UserId = userId

	req := &editFixedRequest{}
	if err := req.bind(c, &editFixed); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if editFixed.SubCategoryName != "" {
		subCategory := model.SubCategoryModel{
			UserNo:          editFixed.UserId,
			CategoryId:      editFixed.CategoryId,
			SubCategoryName: editFixed.SubCategoryName,
		}
		// TODO Createの前に、同じユーザー、同じカテゴリIDに紐づくサブカテゴリ名が存在するか確認
		h.subCategoryStore.CreateSubCategory(&subCategory)
		editFixed.SubCategoryId = subCategory.SubCategoryId
	}

	// err := h.transactionStore.EditFixed(&addTran)
	h.fixedStore.EditFixed(&editFixed)

	return c.JSON(http.StatusOK, "ok")
}
