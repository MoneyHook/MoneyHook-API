package handler

import (
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (h *Handler) getTimelineData(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetTimelineData(userId, month)

	result_list := getTimelineListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getMonthlySpendingData(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlySpendingData(userId, month)

	result_list := getmonthlySpendingDataResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getTransaction(c echo.Context) error {
	userId := getUserId(c)
	transactionId, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		return c.JSON(http.StatusOK, "hej")
	}
	result := h.transactionStore.GetTransactionData(userId, transactionId)

	result_list := getTransactionResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getMonthlyFixedIncome(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyFixedData(userId, month, false)

	result_list := getMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyFixedSpending(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyFixedData(userId, month, true)

	result_list := getMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getHome(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetHome(userId, month)

	result_list := getHomeResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyVariableData(c echo.Context) error {
	userId := getUserId(c)
	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyVariableData(userId, month)

	result_list := getMonthlyVariableResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getTotalSpendingData(c echo.Context) error {
	userId := getUserId(c)
	categoryId := c.QueryParam("category_id")
	subCategoryId := c.QueryParam("sub_category_id")
	startMonth := c.QueryParam("start_month")
	endMonth := c.QueryParam("end_month")

	result := h.transactionStore.GetTotalSpending(userId, categoryId, subCategoryId, startMonth, endMonth)

	result_list := getTotalSpendingResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getFrequentTransactionName(c echo.Context) error {
	userId := getUserId(c)

	result := h.transactionStore.GetFrequentTransactionName(userId)

	result_list := getFrequentTransactionResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) addTransaction(c echo.Context) error {
	userId := getUserId(c)
	var addTran model.AddTransaction

	addTran.UserId = userId

	req := &addTransactionRequest{}
	if err := req.bind(c, &addTran); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("test_error_message")))
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if addTran.SubCategoryId == 0 {
		subCategory := model.SubCategoryModel{
			UserNo:          addTran.UserId,
			CategoryId:      addTran.CategoryId,
			SubCategoryName: addTran.SubCategoryName,
		}
		// TODO Createの前に、同じユーザー、同じカテゴリIDに紐づくサブカテゴリ名が存在するか確認
		h.subCategoryStore.CreateSubCategory(&subCategory)
		addTran.SubCategoryId = subCategory.SubCategoryId
	}

	// err := h.transactionStore.AddTransaction(&addTran)
	h.transactionStore.AddTransaction(&addTran)

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) editTransaction(c echo.Context) error {
	userId := getUserId(c)
	var editTran model.EditTransaction

	editTran.UserId = userId

	req := &editTransactionRequest{}
	if err := req.bind(c, &editTran); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if editTran.SubCategoryId == 0 {
		subCategory := model.SubCategoryModel{
			UserNo:          editTran.UserId,
			CategoryId:      editTran.CategoryId,
			SubCategoryName: editTran.SubCategoryName,
		}
		// TODO Createの前に、同じユーザー、同じカテゴリIDに紐づくサブカテゴリ名が存在するか確認
		h.subCategoryStore.CreateSubCategory(&subCategory)
		editTran.SubCategoryId = subCategory.SubCategoryId
	}

	// err := h.transactionStore.EditTransaction(&addTran)
	h.transactionStore.EditTransaction(&editTran)

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) deleteTransaction(c echo.Context) error {
	userId := getUserId(c)
	transactionId, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		return c.JSON(http.StatusOK, "hej")
	}
	deleteTransaction := model.DeleteTransaction{UserId: userId, TransactionId: transactionId}

	// err := h.transactionStore.EditFixed(&addTran)
	h.transactionStore.DeleteTransaction(&deleteTransaction)

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}
