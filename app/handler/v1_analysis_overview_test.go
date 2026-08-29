package handler

import (
	"MoneyHook/MoneyHook-API/model"
	"testing"
	"time"
)

func TestBuildV1OverviewResponse(t *testing.T) {
	query := v1AnalysisQuery{
		StartDate: "2026-01-01", EndDate: "2026-03-31", GroupBy: "month", Compare: "previous_period",
		Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
	}
	current := []model.V1AnalyticsTransaction{
		{TransactionDate: "2026-01-10", SignedAmount: -1000, FixedFlg: false, CategoryId: "1", CategoryName: "食費"},
		{TransactionDate: "2026-01-20", SignedAmount: 5000, CategoryId: "27", CategoryName: "給与"},
		{TransactionDate: "2026-03-01", SignedAmount: -2000, FixedFlg: true, CategoryId: "22", CategoryName: "住宅"},
	}
	comparison := []model.V1AnalyticsTransaction{
		{TransactionDate: "2025-12-01", SignedAmount: -500, CategoryId: "1", CategoryName: "食費"},
	}
	response, err := buildV1OverviewResponse(query, current, comparison)
	if err != nil {
		t.Fatal(err)
	}
	if response.Summary.ExpenseAmount != 3000 || response.Summary.IncomeAmount != 5000 || response.Summary.NetAmount != 2000 {
		t.Fatalf("unexpected summary: %+v", response.Summary)
	}
	if response.Summary.FixedExpenseAmount != 2000 || response.Summary.VariableExpenseAmount != 1000 {
		t.Fatalf("unexpected fixed/variable summary: %+v", response.Summary)
	}
	if response.Summary.MonthlyAverageExpense != 1000 {
		t.Fatalf("monthly average = %d", response.Summary.MonthlyAverageExpense)
	}
	if len(response.Series) != 3 || response.Series[1].ExpenseAmount != 0 {
		t.Fatalf("series was not zero-filled: %+v", response.Series)
	}
	if len(response.CategoryChanges) != 2 {
		t.Fatalf("category changes = %+v", response.CategoryChanges)
	}
}

func TestAnalysisBucketStartWeek(t *testing.T) {
	value := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	if got := analysisBucketStart(value, "week").Format("2006-01-02"); got != "2026-08-24" {
		t.Fatalf("week bucket = %s", got)
	}
}

func TestBuildV1OverviewResponseWithoutComparisonReturnsNoCategoryChanges(t *testing.T) {
	query := v1AnalysisQuery{
		StartDate: "2026-01-01", EndDate: "2026-01-31", GroupBy: "month", Compare: "none",
		Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
	}
	transactions := []model.V1AnalyticsTransaction{
		{TransactionDate: "2026-01-10", SignedAmount: -1000, CategoryId: "1", CategoryName: "食費"},
	}
	response, err := buildV1OverviewResponse(query, transactions, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.CategoryChanges) != 0 {
		t.Fatalf("category changes = %+v", response.CategoryChanges)
	}
}
