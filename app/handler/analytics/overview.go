package analytics

import (
	"MoneyHook/MoneyHook-API/handler/internal/httpx"
	"MoneyHook/MoneyHook-API/model"
	"math"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
)

type v1AnalysisRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type v1OverviewSummary struct {
	ExpenseAmount         int64 `json:"expense_amount"`
	IncomeAmount          int64 `json:"income_amount"`
	NetAmount             int64 `json:"net_amount"`
	FixedExpenseAmount    int64 `json:"fixed_expense_amount"`
	VariableExpenseAmount int64 `json:"variable_expense_amount"`
	MonthlyAverageExpense int64 `json:"monthly_average_expense"`
}

type v1OverviewSeriesItem struct {
	Bucket                string `json:"bucket"`
	ExpenseAmount         int64  `json:"expense_amount"`
	IncomeAmount          int64  `json:"income_amount"`
	NetAmount             int64  `json:"net_amount"`
	FixedExpenseAmount    int64  `json:"fixed_expense_amount"`
	VariableExpenseAmount int64  `json:"variable_expense_amount"`
}

type v1CategoryChange struct {
	CategoryId       string   `json:"category_id"`
	CategoryName     string   `json:"category_name"`
	CurrentAmount    int64    `json:"current_amount"`
	ComparisonAmount int64    `json:"comparison_amount"`
	DifferenceAmount int64    `json:"difference_amount"`
	DifferenceRate   *float64 `json:"difference_rate"`
}

type v1OverviewResponse struct {
	Range           v1AnalysisRange        `json:"range"`
	Summary         v1OverviewSummary      `json:"summary"`
	Series          []v1OverviewSeriesItem `json:"series"`
	CategoryChanges []v1CategoryChange     `json:"category_changes"`
}

func (h *Handler) GetV1AnalyticsOverview(c echo.Context) error {
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
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "分析データの取得に失敗しました", nil)
	}
	comparison := make([]model.V1AnalyticsTransaction, 0)
	if query.Compare == "previous_period" {
		previousStart, previousEnd := query.previousPeriod()
		comparison, err = h.transactionStore.GetV1AnalyticsTransactions(userId, previousStart, previousEnd)
		if err != nil {
			return httpx.RespondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "比較データの取得に失敗しました", nil)
		}
	}
	response, err := buildV1OverviewResponse(query, transactions, comparison)
	if err != nil {
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INVALID_STORED_DATA", "保存済みデータの日付が不正です", nil)
	}
	return c.JSON(http.StatusOK, response)
}

func buildV1OverviewResponse(query v1AnalysisQuery, transactions []model.V1AnalyticsTransaction, comparison []model.V1AnalyticsTransaction) (v1OverviewResponse, error) {
	response := v1OverviewResponse{
		Range:           v1AnalysisRange{StartDate: query.StartDate, EndDate: query.EndDate},
		Series:          make([]v1OverviewSeriesItem, 0),
		CategoryChanges: make([]v1CategoryChange, 0),
	}
	seriesByBucket := map[string]*v1OverviewSeriesItem{}
	for bucket := analysisBucketStart(query.Start, query.GroupBy); !bucket.After(query.End); bucket = nextAnalysisBucket(bucket, query.GroupBy) {
		key := bucket.Format("2006-01-02")
		item := &v1OverviewSeriesItem{Bucket: key}
		seriesByBucket[key] = item
		response.Series = append(response.Series, *item)
	}

	currentCategories := map[string]*v1CategoryChange{}
	for _, transaction := range transactions {
		date, err := parseStoredTransactionDate(transaction.TransactionDate)
		if err != nil {
			return v1OverviewResponse{}, err
		}
		bucketKey := analysisBucketStart(date, query.GroupBy).Format("2006-01-02")
		item := seriesByBucket[bucketKey]
		if item == nil {
			continue
		}
		applyOverviewAmount(&response.Summary, item, transaction)
		if transaction.SignedAmount < 0 {
			category := currentCategories[transaction.CategoryId]
			if category == nil {
				category = &v1CategoryChange{CategoryId: transaction.CategoryId, CategoryName: transaction.CategoryName}
				currentCategories[transaction.CategoryId] = category
			}
			category.CurrentAmount += -transaction.SignedAmount
		}
	}

	for index := range response.Series {
		if item := seriesByBucket[response.Series[index].Bucket]; item != nil {
			response.Series[index] = *item
		}
	}
	monthCount := (query.End.Year()-query.Start.Year())*12 + int(query.End.Month()-query.Start.Month()) + 1
	response.Summary.MonthlyAverageExpense = int64(math.Round(float64(response.Summary.ExpenseAmount) / float64(monthCount)))
	if query.Compare != "previous_period" {
		return response, nil
	}

	comparisonCategories := map[string]int64{}
	categoryNames := map[string]string{}
	for id, item := range currentCategories {
		categoryNames[id] = item.CategoryName
	}
	for _, transaction := range comparison {
		if transaction.SignedAmount < 0 {
			comparisonCategories[transaction.CategoryId] += -transaction.SignedAmount
			categoryNames[transaction.CategoryId] = transaction.CategoryName
		}
	}
	for id, name := range categoryNames {
		current := int64(0)
		if item := currentCategories[id]; item != nil {
			current = item.CurrentAmount
		}
		comparisonAmount := comparisonCategories[id]
		difference := current - comparisonAmount
		var rate *float64
		if comparisonAmount != 0 {
			value := math.Round((float64(difference)/float64(comparisonAmount))*1000) / 10
			rate = &value
		}
		response.CategoryChanges = append(response.CategoryChanges, v1CategoryChange{
			CategoryId: id, CategoryName: name, CurrentAmount: current,
			ComparisonAmount: comparisonAmount, DifferenceAmount: difference, DifferenceRate: rate,
		})
	}
	sort.Slice(response.CategoryChanges, func(i, j int) bool {
		left := absInt64(response.CategoryChanges[i].DifferenceAmount)
		right := absInt64(response.CategoryChanges[j].DifferenceAmount)
		if left == right {
			return response.CategoryChanges[i].CategoryName < response.CategoryChanges[j].CategoryName
		}
		return left > right
	})
	return response, nil
}

func applyOverviewAmount(summary *v1OverviewSummary, series *v1OverviewSeriesItem, transaction model.V1AnalyticsTransaction) {
	summary.NetAmount += transaction.SignedAmount
	series.NetAmount += transaction.SignedAmount
	if transaction.SignedAmount > 0 {
		summary.IncomeAmount += transaction.SignedAmount
		series.IncomeAmount += transaction.SignedAmount
		return
	}
	expense := -transaction.SignedAmount
	summary.ExpenseAmount += expense
	series.ExpenseAmount += expense
	if transaction.FixedFlg {
		summary.FixedExpenseAmount += expense
		series.FixedExpenseAmount += expense
	} else {
		summary.VariableExpenseAmount += expense
		series.VariableExpenseAmount += expense
	}
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
