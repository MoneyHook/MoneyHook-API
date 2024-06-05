package handler

import (
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (h *Handler) getFixed(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	result := h.fixedStore.GetFixedData(userId)

	result_list := GetFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getDeletedFixed(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	result := h.fixedStore.GetFixedDeletedData(userId)

	result_list := GetFixedDeletedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) addFixed(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var addFixed model.AddFixed

	addFixed.UserId = userId

	req := &addFixedRequest{}
	if err := req.bind(c, &addFixed); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if addFixed.SubCategoryId == 0 {
		subCategory := model.SubCategoryModel{
			UserNo:          addFixed.UserId,
			CategoryId:      addFixed.CategoryId,
			SubCategoryName: addFixed.SubCategoryName,
		}
		// TODO Createの前に、同じユーザー、同じカテゴリIDに紐づくサブカテゴリ名が存在するか確認
		h.subCategoryStore.CreateSubCategory(&subCategory)
		addFixed.SubCategoryId = subCategory.SubCategoryId
	}

	err = h.fixedStore.AddFixed(&addFixed)
	if err != nil {
		log.Printf("database insert error: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) editFixed(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var editFixed model.EditFixed

	editFixed.UserId = userId

	req := &editFixedRequest{}
	if err := req.bind(c, &editFixed); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if editFixed.SubCategoryId == 0 {
		subCategory := model.SubCategoryModel{
			UserNo:          editFixed.UserId,
			CategoryId:      editFixed.CategoryId,
			SubCategoryName: editFixed.SubCategoryName,
		}
		// TODO Createの前に、同じユーザー、同じカテゴリIDに紐づくサブカテゴリ名が存在するか確認
		h.subCategoryStore.CreateSubCategory(&subCategory)
		editFixed.SubCategoryId = subCategory.SubCategoryId
	}

	err = h.fixedStore.EditFixed(&editFixed)
	if err != nil {
		log.Printf("database update error: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("edit_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) deleteFixed(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	monthlyTransactionId, err := strconv.Atoi(c.Param("monthly_transaction_id"))
	if err != nil {
		return c.JSON(http.StatusOK, "hej")
	}
	deleteFixed := model.DeleteFixed{UserId: userId, MonthlyTransactionId: monthlyTransactionId}

	// err := h.transactionStore.EditFixed(&addTran)
	h.fixedStore.DeleteFixed(&deleteFixed)

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}
