package transaction

import (
	"MoneyHook/MoneyHook-API/handler/internal/httpx"
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	subcategorydomain "MoneyHook/MoneyHook-API/subcategory"
	"errors"
	"strconv"

	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	defaultFrequentTransactionLimit = 20
	maxFrequentTransactionLimit     = 100
)

func parseFrequentTransactionLimit(value string) (int, bool) {
	if value == "" {
		return defaultFrequentTransactionLimit, true
	}

	limit, err := strconv.Atoi(value)
	if err != nil || limit < 1 || limit > maxFrequentTransactionLimit {
		return 0, false
	}

	return limit, true
}

func (h *Handler) GetTimelineData(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result, err := h.transactionStore.GetTimelineData(userId, month)
	if err != nil {
		return respondReadError(c, err)
	}

	result_list := GetTimelineListResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) GetMonthlySpendingData(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result, err := h.transactionStore.GetMonthlySpendingData(userId, month)
	if err != nil {
		return respondReadError(c, err)
	}

	result_list := GetmonthlySpendingDataResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) GetTransaction(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	transactionId := c.Param("transactionId")
	result, err := h.transactionStore.GetTransactionData(userId, transactionId)
	if err != nil {
		return respondReadError(c, err)
	}

	if result == nil {
		return c.JSON(http.StatusNotFound, model.Error.Create(message.Get("transaction_not_found")))
	}

	result_list := GetTransactionResponse(result)

	return c.JSON(http.StatusOK, *result_list)
}

func (h *Handler) GetMonthlyFixedIncome(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result, err := h.transactionStore.GetMonthlyFixedData(userId, month, false)
	if err != nil {
		return respondReadError(c, err)
	}

	result_list := GetMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) GetMonthlyFixedSpending(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result, err := h.transactionStore.GetMonthlyFixedData(userId, month, true)
	if err != nil {
		return respondReadError(c, err)
	}

	result_list := GetMonthlyFixedResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) GetHome(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result, err := h.transactionStore.GetHome(userId, month)
	if err != nil {
		return respondReadError(c, err)
	}

	result_list := GetHomeResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) GetMonthlyVariableData(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result, err := h.transactionStore.GetMonthlyVariableData(userId, month)
	if err != nil {
		return respondReadError(c, err)
	}

	result_list := GetMonthlyVariableResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) GetTotalSpendingData(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	categoryId := c.QueryParam("category_id")
	subCategoryId := c.QueryParam("sub_category_id")
	startMonth := c.QueryParam("start_month")
	endMonth := c.QueryParam("end_month")

	result, err := h.transactionStore.GetTotalSpending(userId, categoryId, subCategoryId, startMonth, endMonth)
	if err != nil {
		return respondReadError(c, err)
	}

	result_list := GetTotalSpendingResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) GroupByPayment(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	month := c.QueryParam("month")

	result, err := h.transactionStore.GetGroupByPayment(userId, month)
	if err != nil {
		return respondReadError(c, err)
	}

	last_month, err := time.Parse("2006-01-02", month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("date_parse_error")))
	}

	last_month_result, err := h.transactionStore.GetLastMonthGroupByPayment(userId, last_month.AddDate(0, -1, 0).Format("2006-01-02"))
	if err != nil {
		return respondReadError(c, err)
	}

	result_list := GetPaymentGroupResponse(result, last_month_result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) GetMonthlyWithdrawalAmount(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	str_month := c.QueryParam("month")

	month, err := time.Parse("2006-01-02", str_month)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("date_parse_error")))
	}

	var result []*model.MonthlyWithdrawalAmountList

	payment_list, err := h.paymentResourceStore.GetPaymentResourceList(userId)
	if err != nil {
		return respondReadError(c, err)
	}
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

			monthlyWithdrawalAmount, err := h.transactionStore.GetMonthlyWithdrawalAmount(userId, payment.PaymentId, startMonth.Format("2006-01-02"), endMonth.Format("2006-01-02"))
			if err != nil {
				return respondReadError(c, err)
			}
			if monthlyWithdrawalAmount.PaymentId != "" {
				result = append(result, monthlyWithdrawalAmount)
			}
		}
	}

	result_list := GetMonthlyWithdrawalAmount(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) GetFrequentTransactionName(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	limit, valid := parseFrequentTransactionLimit(c.QueryParam("limit"))
	if !valid {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"status":  "error",
			"message": "limitは1〜100の整数で指定してください。",
		})
	}

	result, err := h.transactionStore.GetFrequentTransactionName(userId, limit)
	if err != nil {
		return respondReadError(c, err)
	}

	result_list := GetFrequentTransactionResponse(result)

	return c.JSON(http.StatusOK, result_list)
}

func (h *Handler) AddTransaction(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var addTran model.AddTransaction

	addTran.UserId = userId

	req := &AddTransactionRequest{}
	if err := req.Bind(c, &addTran); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	err = h.transactionStore.AddTransaction(&addTran)
	if err != nil {
		if errors.Is(err, subcategorydomain.ErrResolveFailed) {
			return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("sub_category_create_failed")))
		}
		log.Printf("AddTransaction: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) AddTransactionList(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var addTranList model.AddTransactionList

	addTranList.UserId = userId

	req := &AddTransactionListRequest{}
	if err := req.Bind(c, &addTranList); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	err = h.transactionStore.AddTransactionList(&addTranList)
	if err != nil {
		if errors.Is(err, subcategorydomain.ErrResolveFailed) {
			return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("sub_category_create_failed")))
		}
		log.Printf("AddTransactionList: %v\n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("add_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) EditTransaction(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.Error.Create(message.Get("token_expired_error")))
	}

	var editTran model.EditTransaction

	editTran.UserId = userId

	req := &EditTransactionRequest{}
	if err := req.Bind(c, &editTran); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, "error")
		// return c.JSON(http.StatusUnprocessableEntity, err)
	}

	err = h.transactionStore.EditTransaction(&editTran)
	if err != nil {
		if errors.Is(err, subcategorydomain.ErrResolveFailed) {
			return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("sub_category_create_failed")))
		}
		log.Printf("EditTransaction: %v/n", err)
		return c.JSON(http.StatusUnprocessableEntity, model.Error.Create(message.Get("edit_failed")))
	}

	return c.JSON(http.StatusOK, model.Success.Create(nil))
}

func (h *Handler) DeleteTransaction(c echo.Context) error {
	userId, err := httpx.UserID(c)
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

func respondReadError(c echo.Context, err error) error {
	c.Logger().Error(err)
	return httpx.RespondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "データの取得に失敗しました", nil)
}
