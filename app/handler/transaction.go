package handler

import (
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func (h *Handler) getTimelineData(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetTimelineData(userId, month)

	result_list := getTimelineListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getMonthlySpendingData(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlySpendingData(userId, month)

	result_list := getmonthlySpendingDataResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getTransaction(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	transactionId, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		return c.JSON(http.StatusOK, "hej")
	}
	result := h.transactionStore.GetTransactionData(userId, transactionId)

	result_list := getTransactionResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getMonthlyFixedIncome(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyFixedData(userId, month, false)

	result_list := getMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyFixedSpending(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyFixedData(userId, month, true)

	result_list := getMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getHome(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetHome(userId, month)

	result_list := getHomeResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyVariableData(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyVariableData(userId, month)

	result_list := getMonthlyVariableResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getTotalSpendingData(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	categoryId := c.QueryParam("category_id")
	subCategoryId := c.QueryParam("sub_category_id")
	startMonth := c.QueryParam("start_month")
	endMonth := c.QueryParam("end_month")

	result := h.transactionStore.GetTotalSpending(userId, categoryId, subCategoryId, startMonth, endMonth)

	result_list := getTotalSpendingResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) groupByPayment(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetGroupByPayment(userId, month)

	last_month, err := time.Parse("2006-01-02", month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("date_parse_error")))
	}

	last_month_result := h.transactionStore.GetLastMonthGroupByPayment(userId, last_month.AddDate(0, -1, 0).Format("2006-01-02"))

	result_list := getPaymentGroupResponse(result, last_month_result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyWithdrawalAmount(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	last_month, err := time.Parse("2006-01-02", month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("date_parse_error")))
	}

	result := h.transactionStore.GetMonthlyWithdrawalAmount(userId, last_month.AddDate(0, -1, 0).Format("2006-01-02"))

	result_list := getMonthlyWithdrawalAmount(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getFrequentTransactionName(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	result := h.transactionStore.GetFrequentTransactionName(userId)

	result_list := getFrequentTransactionResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) addTransaction(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

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
		// Createの前に、同じユーザー、同じカテゴリIDに紐づくサブカテゴリ名が存在するか確認
		if !h.subCategoryStore.FindByName(&subCategory) {
			// サブカテゴリ作成
			error := h.subCategoryStore.CreateSubCategory(&subCategory)
			if error != nil {
				log.Printf("database insert error: %v\n", err)
				log.Printf("'%v' is exist: %v\n", subCategory.SubCategoryName, subCategory.SubCategoryId != 0)
				return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("sub_category_create_failed")))
			}
		}

		addTran.SubCategoryId = subCategory.SubCategoryId
	}

	err = h.transactionStore.AddTransaction(&addTran)
	if err != nil {
		log.Printf("AddTransaction: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) addTransactionList(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var addTranList model.AddTransactionList

	addTranList.UserId = userId

	req := &addTransactionListRequest{}
	if err := req.bind(c, &addTranList); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("test_error_message")))
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	for i, addTran := range addTranList.TransactionList {
		if addTran.SubCategoryId == 0 {
			subCategory := model.SubCategoryModel{
				UserNo:          addTranList.UserId,
				CategoryId:      addTran.CategoryId,
				SubCategoryName: addTran.SubCategoryName,
			}
			// Createの前に、同じユーザー、同じカテゴリIDに紐づくサブカテゴリ名が存在するか確認
			if !h.subCategoryStore.FindByName(&subCategory) {
				// サブカテゴリ作成
				error := h.subCategoryStore.CreateSubCategory(&subCategory)
				if error != nil {
					log.Printf("CreateSubCategory: %v\n", error)
					log.Printf("'%v' is exist: %v\n", subCategory.SubCategoryName, subCategory.SubCategoryId != 0)
					return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("sub_category_create_failed")))
				}
			}

			addTranList.TransactionList[i].SubCategoryId = subCategory.SubCategoryId
		}
	}

	err = h.transactionStore.AddTransactionList(&addTranList)
	if err != nil {
		log.Printf("AddTransactionList: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) editTransaction(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

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

	err = h.transactionStore.EditTransaction(&editTran)
	if err != nil {
		log.Printf("EditTransaction: %v/n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("edit_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) deleteTransaction(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	transactionId, err := strconv.Atoi(c.Param("transactionId"))
	if err != nil {
		return c.JSON(http.StatusOK, "hej")
	}
	deleteTransaction := model.DeleteTransaction{UserId: userId, TransactionId: transactionId}

	err = h.transactionStore.DeleteTransaction(&deleteTransaction)
	if err != nil {
		log.Printf("DeleteTransaction: %v/n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("delete_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}
