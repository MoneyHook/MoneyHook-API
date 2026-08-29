package handler

import (
	"MoneyHook/MoneyHook-API/model"
	"math"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
)

const unclassifiedPaymentKey = "__unclassified__"

type v1PaymentAnalysis struct {
	PaymentId         *string                          `json:"payment_id"`
	PaymentName       string                           `json:"payment_name"`
	PaymentTypeId     *string                          `json:"payment_type_id"`
	PaymentTypeName   *string                          `json:"payment_type_name"`
	IsPaymentDueLater *bool                            `json:"is_payment_due_later"`
	ExpenseAmount     int64                            `json:"expense_amount"`
	Ratio             float64                          `json:"ratio"`
	TransactionCount  int                              `json:"transaction_count"`
	AverageAmount     int64                            `json:"average_amount"`
	Series            []v1ExpenseSeriesItem            `json:"series"`
	TransactionList   []v1AnalyticsTransactionResource `json:"transaction_list"`
	seriesByBucket    map[string]*v1ExpenseSeriesItem
}

type v1PaymentsResponse struct {
	Range              v1AnalysisRange     `json:"range"`
	TotalExpenseAmount int64               `json:"total_expense_amount"`
	PaymentList        []v1PaymentAnalysis `json:"payment_list"`
}

func (h *Handler) getV1AnalyticsPayments(c echo.Context) error {
	userId, err := h.GetV1UserId(c)
	if err != nil {
		return respondV1Unauthorized(c)
	}
	query, fieldErrors := parseV1AnalysisQuery(c)
	if len(fieldErrors) > 0 {
		return respondV1Error(c, http.StatusBadRequest, "INVALID_QUERY", "分析条件を確認してください", fieldErrors)
	}
	transactions, err := h.transactionStore.GetV1AnalyticsTransactions(userId, query.StartDate, query.EndDate)
	if err != nil {
		return respondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "支払方法分析データの取得に失敗しました", nil)
	}
	response, err := buildV1PaymentsResponse(query, transactions)
	if err != nil {
		return respondV1Error(c, http.StatusInternalServerError, "INVALID_STORED_DATA", "保存済みデータの日付が不正です", nil)
	}
	return c.JSON(http.StatusOK, response)
}

func buildV1PaymentsResponse(query v1AnalysisQuery, transactions []model.V1AnalyticsTransaction) (v1PaymentsResponse, error) {
	response := v1PaymentsResponse{
		Range:       v1AnalysisRange{StartDate: query.StartDate, EndDate: query.EndDate},
		PaymentList: make([]v1PaymentAnalysis, 0),
	}
	payments := map[string]*v1PaymentAnalysis{}
	for _, transaction := range transactions {
		if transaction.SignedAmount >= 0 {
			continue
		}
		date, err := parseStoredTransactionDate(transaction.TransactionDate)
		if err != nil {
			return v1PaymentsResponse{}, err
		}
		bucket := analysisBucketStart(date, query.GroupBy).Format("2006-01-02")
		key := unclassifiedPaymentKey
		name := "未分類"
		if transaction.PaymentId != nil {
			key = *transaction.PaymentId
			if transaction.PaymentName != nil && *transaction.PaymentName != "" {
				name = *transaction.PaymentName
			}
		}
		payment := payments[key]
		if payment == nil {
			payment = &v1PaymentAnalysis{
				PaymentId: transaction.PaymentId, PaymentName: name,
				PaymentTypeId: transaction.PaymentTypeId, PaymentTypeName: transaction.PaymentTypeName,
				IsPaymentDueLater: transaction.IsPaymentDueLater,
				Series:            emptyExpenseSeries(query), TransactionList: make([]v1AnalyticsTransactionResource, 0),
			}
			payment.seriesByBucket = indexExpenseSeries(payment.Series)
			payments[key] = payment
		}
		expense := -transaction.SignedAmount
		response.TotalExpenseAmount += expense
		payment.ExpenseAmount += expense
		payment.TransactionCount++
		payment.seriesByBucket[bucket].ExpenseAmount += expense
		payment.TransactionList = append(payment.TransactionList, newV1AnalyticsTransactionResource(transaction))
	}

	for _, payment := range payments {
		payment.Ratio = percentage(payment.ExpenseAmount, response.TotalExpenseAmount)
		payment.AverageAmount = int64(math.Round(float64(payment.ExpenseAmount) / float64(payment.TransactionCount)))
		payment.Series = materializeExpenseSeries(payment.Series, payment.seriesByBucket)
		payment.seriesByBucket = nil
		response.PaymentList = append(response.PaymentList, *payment)
	}
	sort.Slice(response.PaymentList, func(i, j int) bool {
		if response.PaymentList[i].ExpenseAmount == response.PaymentList[j].ExpenseAmount {
			return response.PaymentList[i].PaymentName < response.PaymentList[j].PaymentName
		}
		return response.PaymentList[i].ExpenseAmount > response.PaymentList[j].ExpenseAmount
	})
	return response, nil
}
