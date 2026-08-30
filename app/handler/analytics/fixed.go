package analytics

import (
	"MoneyHook/MoneyHook-API/handler/internal/httpx"
	"MoneyHook/MoneyHook-API/model"
	"math"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
)

type v1FixedSummary struct {
	ExpenseAmount        int64    `json:"expense_amount"`
	MonthlyAverage       int64    `json:"monthly_average"`
	AnnualizedAmount     int64    `json:"annualized_amount"`
	TotalExpenseRatio    float64  `json:"total_expense_ratio"`
	LatestBucketAmount   int64    `json:"latest_bucket_amount"`
	PreviousBucketAmount int64    `json:"previous_bucket_amount"`
	DifferenceAmount     int64    `json:"difference_amount"`
	DifferenceRate       *float64 `json:"difference_rate"`
}

type v1FixedCategory struct {
	CategoryId       string                           `json:"category_id"`
	CategoryName     string                           `json:"category_name"`
	ExpenseAmount    int64                            `json:"expense_amount"`
	Ratio            float64                          `json:"ratio"`
	MonthlyAverage   int64                            `json:"monthly_average"`
	AnnualizedAmount int64                            `json:"annualized_amount"`
	Series           []v1ExpenseSeriesItem            `json:"series"`
	TransactionList  []v1AnalyticsTransactionResource `json:"transaction_list"`
	seriesByBucket   map[string]*v1ExpenseSeriesItem
}

type v1FixedResponse struct {
	Range        v1AnalysisRange       `json:"range"`
	Summary      v1FixedSummary        `json:"summary"`
	Series       []v1ExpenseSeriesItem `json:"series"`
	CategoryList []v1FixedCategory     `json:"category_list"`
}

func (h *Handler) GetV1AnalyticsFixed(c echo.Context) error {
	userId, err := httpx.UserID(c)
	if err != nil {
		return httpx.RespondV1Unauthorized(c)
	}
	query, fieldErrors := parseV1AnalysisQuery(c)
	if len(fieldErrors) > 0 {
		return httpx.RespondV1Error(c, http.StatusBadRequest, "INVALID_QUERY", "分析条件を確認してください", fieldErrors)
	}
	transactions, err := h.transactionStore.GetV1AnalyticsTransactions(userId, query.StartDate, query.EndDate)
	if err != nil {
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "固定費分析データの取得に失敗しました", nil)
	}
	response, err := buildV1FixedResponse(query, transactions)
	if err != nil {
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INVALID_STORED_DATA", "保存済みデータの日付が不正です", nil)
	}
	return c.JSON(http.StatusOK, response)
}

func buildV1FixedResponse(query v1AnalysisQuery, transactions []model.V1AnalyticsTransaction) (v1FixedResponse, error) {
	response := v1FixedResponse{
		Range:  v1AnalysisRange{StartDate: query.StartDate, EndDate: query.EndDate},
		Series: emptyExpenseSeries(query), CategoryList: make([]v1FixedCategory, 0),
	}
	seriesByBucket := indexExpenseSeries(response.Series)
	categories := map[string]*v1FixedCategory{}
	totalExpense := int64(0)
	monthCount := (query.End.Year()-query.Start.Year())*12 + int(query.End.Month()-query.Start.Month()) + 1

	for _, transaction := range transactions {
		if transaction.SignedAmount >= 0 {
			continue
		}
		expense := -transaction.SignedAmount
		totalExpense += expense
		if !transaction.FixedFlg {
			continue
		}
		date, err := parseStoredTransactionDate(transaction.TransactionDate)
		if err != nil {
			return v1FixedResponse{}, err
		}
		bucket := analysisBucketStart(date, query.GroupBy).Format("2006-01-02")
		response.Summary.ExpenseAmount += expense
		seriesByBucket[bucket].ExpenseAmount += expense
		category := categories[transaction.CategoryId]
		if category == nil {
			category = &v1FixedCategory{
				CategoryId: transaction.CategoryId, CategoryName: transaction.CategoryName,
				Series: emptyExpenseSeries(query), TransactionList: make([]v1AnalyticsTransactionResource, 0),
			}
			category.seriesByBucket = indexExpenseSeries(category.Series)
			categories[transaction.CategoryId] = category
		}
		category.ExpenseAmount += expense
		category.seriesByBucket[bucket].ExpenseAmount += expense
		category.TransactionList = append(category.TransactionList, newV1AnalyticsTransactionResource(transaction))
	}

	response.Series = materializeExpenseSeries(response.Series, seriesByBucket)
	response.Summary.MonthlyAverage = int64(math.Round(float64(response.Summary.ExpenseAmount) / float64(monthCount)))
	response.Summary.AnnualizedAmount = response.Summary.MonthlyAverage * 12
	response.Summary.TotalExpenseRatio = percentage(response.Summary.ExpenseAmount, totalExpense)
	if len(response.Series) > 0 {
		response.Summary.LatestBucketAmount = response.Series[len(response.Series)-1].ExpenseAmount
	}
	if len(response.Series) > 1 {
		response.Summary.PreviousBucketAmount = response.Series[len(response.Series)-2].ExpenseAmount
	}
	response.Summary.DifferenceAmount = response.Summary.LatestBucketAmount - response.Summary.PreviousBucketAmount
	if response.Summary.PreviousBucketAmount != 0 {
		rate := math.Round((float64(response.Summary.DifferenceAmount)/float64(response.Summary.PreviousBucketAmount))*1000) / 10
		response.Summary.DifferenceRate = &rate
	}

	for _, category := range categories {
		category.Series = materializeExpenseSeries(category.Series, category.seriesByBucket)
		category.Ratio = percentage(category.ExpenseAmount, response.Summary.ExpenseAmount)
		category.MonthlyAverage = int64(math.Round(float64(category.ExpenseAmount) / float64(monthCount)))
		category.AnnualizedAmount = category.MonthlyAverage * 12
		category.seriesByBucket = nil
		response.CategoryList = append(response.CategoryList, *category)
	}
	sort.Slice(response.CategoryList, func(i, j int) bool {
		if response.CategoryList[i].ExpenseAmount == response.CategoryList[j].ExpenseAmount {
			return response.CategoryList[i].CategoryName < response.CategoryList[j].CategoryName
		}
		return response.CategoryList[i].ExpenseAmount > response.CategoryList[j].ExpenseAmount
	})
	return response, nil
}
