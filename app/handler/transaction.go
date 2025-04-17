package handler

import (
	"MoneyHook/MoneyHook-API/handler/request"
	"MoneyHook/MoneyHook-API/handler/response"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"

	"log"
	"net/http"
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

	result_list := response.GetTimelineListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getMonthlySpendingData(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlySpendingData(userId, month)

	result_list := response.GetmonthlySpendingDataResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getTransaction(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	transactionId := c.Param("transactionId")
	result := h.transactionStore.GetTransactionData(userId, transactionId)

	if result == nil {
		return c.JSON(http.StatusNotFound, model.Error.Create(message.Get("transaction_not_found")))
	}

	result_list := response.GetTransactionResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) getMonthlyFixedIncome(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyFixedData(userId, month, false)

	result_list := response.GetMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyFixedSpending(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyFixedData(userId, month, true)

	result_list := response.GetMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getHome(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetHome(userId, month)

	result_list := response.GetHomeResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyVariableData(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result := h.transactionStore.GetMonthlyVariableData(userId, month)

	result_list := response.GetMonthlyVariableResponse(result)

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

	result_list := response.GetTotalSpendingResponse(result)

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

	result_list := response.GetPaymentGroupResponse(result, last_month_result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getMonthlyWithdrawalAmount(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	str_month := c.QueryParam("month")

	month, err := time.Parse("2006-01-02", str_month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("date_parse_error")))
	}

	var result []*model.MonthlyWithdrawalAmountList

	payment_list := h.paymentResourceStore.GetPaymentResourceList(userId)
	for _, payment := range *payment_list {
		if payment.PaymentDate != 0 {
			var startMonth time.Time
			var endMonth time.Time

			if month.AddDate(0, 0, -1).Day() <= payment.ClosingDate {
				/*
					前月の末日 <= 登録した締日 の場合、前月の初日から前月末までが対象
					例
					締日        : 31日
					前月の末尾   : 29日(2024-02-29)
					startMonth : 2024-02-01
					endMonth   : 2024-02-29
				*/
				startMonth = month.AddDate(0, -1, 0)
				endMonth = month.AddDate(0, 0, -1)
			} else {
				/*
					上記以外の場合、「前々月の締日+1日」から「前月の締日」までが対象
					例
					締日        : 10日
					startMonth : 2024-01-11
					endMonth   : 2024-02-10
				*/
				startMonth = month.AddDate(0, -2, payment.ClosingDate)
				endMonth = month.AddDate(0, -1, payment.ClosingDate-1)
			}

			monthlyWithdrawalAmount := h.transactionStore.GetMonthlyWithdrawalAmount(userId, payment.PaymentId, startMonth.Format("2006-01-02"), endMonth.Format("2006-01-02"))
			if monthlyWithdrawalAmount.PaymentId != "" {
				result = append(result, monthlyWithdrawalAmount)
			}
		}
	}

	result_list := response.GetMonthlyWithdrawalAmount(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) getFrequentTransactionName(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	result := h.transactionStore.GetFrequentTransactionName(userId)

	result_list := response.GetFrequentTransactionResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) addTransaction(c echo.Context) error {
	userId, err := h.GetUserId(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var addTran model.AddTransaction

	addTran.UserId = userId

	req := &request.AddTransactionRequest{}
	if err := req.Bind(c, &addTran); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if addTran.SubCategoryId == "" {
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
				log.Printf("'%v' is exist: %v\n", subCategory.SubCategoryName, subCategory.SubCategoryId != "")
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

	req := &request.AddTransactionListRequest{}
	if err := req.Bind(c, &addTranList); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	for i, addTran := range addTranList.TransactionList {
		if addTran.SubCategoryId == "" {
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
					log.Printf("'%v' is exist: %v\n", subCategory.SubCategoryName, subCategory.SubCategoryId != "")
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

	req := &request.EditTransactionRequest{}
	if err := req.Bind(c, &editTran); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	if editTran.SubCategoryId == "" {
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

	transactionId := c.Param("transactionId")
	deleteTransaction := model.DeleteTransaction{UserId: userId, TransactionId: transactionId}

	err = h.transactionStore.DeleteTransaction(&deleteTransaction)
	if err != nil {
		log.Printf("DeleteTransaction: %v/n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("delete_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}
