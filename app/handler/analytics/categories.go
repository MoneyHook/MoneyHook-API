package analytics

import (
	"MoneyHook/MoneyHook-API/handler/internal/httpx"
	"MoneyHook/MoneyHook-API/model"
	"math"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
)

type v1ExpenseSeriesItem struct {
	Bucket        string `json:"bucket"`
	ExpenseAmount int64  `json:"expense_amount"`
}

type v1AnalyticsTransactionResource struct {
	TransactionId   string  `json:"transaction_id"`
	TransactionDate string  `json:"transaction_date"`
	TransactionTime *string `json:"transaction_time"`
	TransactionName string  `json:"transaction_name"`
	Amount          int64   `json:"amount"`
	Sign            int     `json:"sign"`
	SignedAmount    int64   `json:"signed_amount"`
	CategoryId      string  `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	SubCategoryId   string  `json:"sub_category_id"`
	SubCategoryName string  `json:"sub_category_name"`
	FixedFlg        bool    `json:"fixed_flg"`
	PaymentId       *string `json:"payment_id"`
	PaymentName     *string `json:"payment_name"`
}

type v1SubCategoryAnalysis struct {
	SubCategoryId   string                           `json:"sub_category_id"`
	SubCategoryName string                           `json:"sub_category_name"`
	ExpenseAmount   int64                            `json:"expense_amount"`
	Ratio           float64                          `json:"ratio"`
	Series          []v1ExpenseSeriesItem            `json:"series"`
	TransactionList []v1AnalyticsTransactionResource `json:"transaction_list"`
	seriesByBucket  map[string]*v1ExpenseSeriesItem
}

type v1CategoryAnalysis struct {
	CategoryId      string                           `json:"category_id"`
	CategoryName    string                           `json:"category_name"`
	ExpenseAmount   int64                            `json:"expense_amount"`
	Ratio           float64                          `json:"ratio"`
	Series          []v1ExpenseSeriesItem            `json:"series"`
	SubCategoryList []v1SubCategoryAnalysis          `json:"sub_category_list"`
	TransactionList []v1AnalyticsTransactionResource `json:"transaction_list"`
	seriesByBucket  map[string]*v1ExpenseSeriesItem
	subCategories   map[string]*v1SubCategoryAnalysis
}

type v1CategoriesResponse struct {
	Range              v1AnalysisRange      `json:"range"`
	TotalExpenseAmount int64                `json:"total_expense_amount"`
	CategoryList       []v1CategoryAnalysis `json:"category_list"`
}

func (h *Handler) GetV1AnalyticsCategories(c echo.Context) error {
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
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "カテゴリ分析データの取得に失敗しました", nil)
	}
	response, err := buildV1CategoriesResponse(query, transactions)
	if err != nil {
		return httpx.RespondV1Error(c, http.StatusInternalServerError, "INVALID_STORED_DATA", "保存済みデータの日付が不正です", nil)
	}
	return c.JSON(http.StatusOK, response)
}

func buildV1CategoriesResponse(query v1AnalysisQuery, transactions []model.V1AnalyticsTransaction) (v1CategoriesResponse, error) {
	response := v1CategoriesResponse{
		Range:        v1AnalysisRange{StartDate: query.StartDate, EndDate: query.EndDate},
		CategoryList: make([]v1CategoryAnalysis, 0),
	}
	categories := map[string]*v1CategoryAnalysis{}
	for _, transaction := range transactions {
		if transaction.SignedAmount >= 0 {
			continue
		}
		date, err := parseStoredTransactionDate(transaction.TransactionDate)
		if err != nil {
			return v1CategoriesResponse{}, err
		}
		bucket := analysisBucketStart(date, query.GroupBy).Format("2006-01-02")
		category := categories[transaction.CategoryId]
		if category == nil {
			category = &v1CategoryAnalysis{
				CategoryId: transaction.CategoryId, CategoryName: transaction.CategoryName,
				Series: emptyExpenseSeries(query), SubCategoryList: make([]v1SubCategoryAnalysis, 0),
				TransactionList: make([]v1AnalyticsTransactionResource, 0),
				subCategories:   map[string]*v1SubCategoryAnalysis{},
			}
			category.seriesByBucket = indexExpenseSeries(category.Series)
			categories[transaction.CategoryId] = category
		}
		subCategory := category.subCategories[transaction.SubCategoryId]
		if subCategory == nil {
			subCategory = &v1SubCategoryAnalysis{
				SubCategoryId: transaction.SubCategoryId, SubCategoryName: transaction.SubCategoryName,
				Series: emptyExpenseSeries(query), TransactionList: make([]v1AnalyticsTransactionResource, 0),
			}
			subCategory.seriesByBucket = indexExpenseSeries(subCategory.Series)
			category.subCategories[transaction.SubCategoryId] = subCategory
		}
		expense := -transaction.SignedAmount
		response.TotalExpenseAmount += expense
		category.ExpenseAmount += expense
		subCategory.ExpenseAmount += expense
		category.seriesByBucket[bucket].ExpenseAmount += expense
		subCategory.seriesByBucket[bucket].ExpenseAmount += expense
		resource := newV1AnalyticsTransactionResource(transaction)
		category.TransactionList = append(category.TransactionList, resource)
		subCategory.TransactionList = append(subCategory.TransactionList, resource)
	}

	for _, category := range categories {
		category.Series = materializeExpenseSeries(category.Series, category.seriesByBucket)
		category.Ratio = percentage(category.ExpenseAmount, response.TotalExpenseAmount)
		for _, subCategory := range category.subCategories {
			subCategory.Series = materializeExpenseSeries(subCategory.Series, subCategory.seriesByBucket)
			subCategory.Ratio = percentage(subCategory.ExpenseAmount, category.ExpenseAmount)
			category.SubCategoryList = append(category.SubCategoryList, *subCategory)
		}
		sort.Slice(category.SubCategoryList, func(i, j int) bool {
			if category.SubCategoryList[i].ExpenseAmount == category.SubCategoryList[j].ExpenseAmount {
				return category.SubCategoryList[i].SubCategoryName < category.SubCategoryList[j].SubCategoryName
			}
			return category.SubCategoryList[i].ExpenseAmount > category.SubCategoryList[j].ExpenseAmount
		})
		category.seriesByBucket = nil
		category.subCategories = nil
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

func emptyExpenseSeries(query v1AnalysisQuery) []v1ExpenseSeriesItem {
	series := make([]v1ExpenseSeriesItem, 0)
	for bucket := analysisBucketStart(query.Start, query.GroupBy); !bucket.After(query.End); bucket = nextAnalysisBucket(bucket, query.GroupBy) {
		series = append(series, v1ExpenseSeriesItem{Bucket: bucket.Format("2006-01-02")})
	}
	return series
}

func indexExpenseSeries(series []v1ExpenseSeriesItem) map[string]*v1ExpenseSeriesItem {
	result := make(map[string]*v1ExpenseSeriesItem, len(series))
	for index := range series {
		result[series[index].Bucket] = &series[index]
	}
	return result
}

func materializeExpenseSeries(series []v1ExpenseSeriesItem, indexed map[string]*v1ExpenseSeriesItem) []v1ExpenseSeriesItem {
	for index := range series {
		series[index] = *indexed[series[index].Bucket]
	}
	return series
}

func percentage(part int64, total int64) float64 {
	if total == 0 {
		return 0
	}
	return math.Round((float64(part)/float64(total))*1000) / 10
}

func newV1AnalyticsTransactionResource(transaction model.V1AnalyticsTransaction) v1AnalyticsTransactionResource {
	sign := 1
	amount := transaction.SignedAmount
	if amount < 0 {
		sign = -1
		amount = -amount
	}
	return v1AnalyticsTransactionResource{
		TransactionId: transaction.TransactionId, TransactionDate: transaction.TransactionDate,
		TransactionTime: transaction.TransactionTime, TransactionName: transaction.TransactionName,
		Amount: amount, Sign: sign, SignedAmount: transaction.SignedAmount,
		CategoryId: transaction.CategoryId, CategoryName: transaction.CategoryName,
		SubCategoryId: transaction.SubCategoryId, SubCategoryName: transaction.SubCategoryName,
		FixedFlg: transaction.FixedFlg, PaymentId: transaction.PaymentId, PaymentName: transaction.PaymentName,
	}
}
