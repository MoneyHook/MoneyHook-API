package analytics

import (
	"MoneyHook/MoneyHook-API/model"
	"testing"
	"time"
)

func TestBuildV1CategoriesResponseKeepsExpenseHierarchyConsistent(t *testing.T) {
	query := v1AnalysisQuery{
		StartDate: "2026-01-01", EndDate: "2026-02-28", GroupBy: "month",
		Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
	}
	transactions := []model.V1AnalyticsTransaction{
		{TransactionId: "1", TransactionDate: "2026-01-10", SignedAmount: -1000, CategoryId: "1", CategoryName: "食費", SubCategoryId: "11", SubCategoryName: "外食"},
		{TransactionId: "2", TransactionDate: "2026-02-10", SignedAmount: -500, CategoryId: "1", CategoryName: "食費", SubCategoryId: "12", SubCategoryName: "スーパー"},
		{TransactionId: "3", TransactionDate: "2026-02-11", SignedAmount: 300, CategoryId: "1", CategoryName: "食費", SubCategoryId: "11", SubCategoryName: "外食"},
	}
	response, err := buildV1CategoriesResponse(query, transactions)
	if err != nil {
		t.Fatal(err)
	}
	if response.TotalExpenseAmount != 1500 || len(response.CategoryList) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
	category := response.CategoryList[0]
	if category.ExpenseAmount != 1500 || len(category.TransactionList) != 2 || len(category.SubCategoryList) != 2 {
		t.Fatalf("unexpected category: %+v", category)
	}
	var subTotal int64
	for _, subCategory := range category.SubCategoryList {
		subTotal += subCategory.ExpenseAmount
		if len(subCategory.Series) != 2 {
			t.Fatalf("subcategory series was not zero-filled: %+v", subCategory.Series)
		}
	}
	if subTotal != category.ExpenseAmount {
		t.Fatalf("subcategory total %d != category total %d", subTotal, category.ExpenseAmount)
	}
}
