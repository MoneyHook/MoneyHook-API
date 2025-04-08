package handler

import (
	"MoneyHook/MoneyHook-API/handler/request"
	"MoneyHook/MoneyHook-API/handler/response"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetSubCategoryList(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	categoryId := c.Param("categoryId")

	result := h.subCategoryStore.GetSubCategoryList(userId, categoryId)

	result_list := response.GetSubCategoryListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) EditSubCategory(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var editSubCategory model.EditSubCategoryModel

	editSubCategory.UserId = userId

	req := &request.EditSubCategoryRequest{}
	if err := req.Bind(c, &editSubCategory); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if editSubCategory.IsEnable {
		err = h.subCategoryStore.ExposeSubCategory(&editSubCategory)
		if err != nil {
			log.Printf("ExposeSubCategory: %v/n", err)
			return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("delete_failed")))
		}
	} else {
		err = h.subCategoryStore.HideSubCategory(&editSubCategory)
		if err != nil {
			log.Printf("HideSubCategory: %v/n", err)
			return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
		}
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}
