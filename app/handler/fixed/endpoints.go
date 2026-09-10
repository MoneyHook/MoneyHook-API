package fixed

import (
	"MoneyHook/MoneyHook-API/handler/internal/httpx"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	subcategorydomain "MoneyHook/MoneyHook-API/subcategory"
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetFixed(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	result := h.fixedStore.GetFixedData(userId)

	result_list := GetFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) GetDeletedFixed(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	result := h.fixedStore.GetFixedDeletedData(userId)

	result_list := GetFixedDeletedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) AddFixed(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var addFixed model.AddFixed

	addFixed.UserId = userId

	req := &AddFixedRequest{}
	if err := req.Bind(c, &addFixed); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	err = h.fixedStore.AddFixed(&addFixed)
	if err != nil {
		if errors.Is(err, subcategorydomain.ErrResolveFailed) {
			return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("sub_category_create_failed")))
		}
		log.Printf("AddFixed: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) EditFixed(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var editFixed model.EditFixed

	editFixed.UserId = userId

	req := &EditFixedRequest{}
	if err := req.Bind(c, &editFixed); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	err = h.fixedStore.EditFixed(&editFixed)
	if err != nil {
		if errors.Is(err, subcategorydomain.ErrResolveFailed) {
			return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("sub_category_create_failed")))
		}
		log.Printf("EditFixed: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("edit_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) DeleteFixed(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	monthlyTransactionId := c.Param("monthly_transaction_id")
	deleteFixed := model.DeleteFixed{UserId: userId, MonthlyTransactionId: monthlyTransactionId}

	err = h.fixedStore.DeleteFixed(&deleteFixed)
	if err != nil {
		log.Printf("DeleteFixed: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("delete_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}
