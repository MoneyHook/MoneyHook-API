package handler

import (
	"MoneyHook/MoneyHook-API/model"
	"testing"
	"time"
)

func TestBuildV1FixedResponseUsesOnlyFixedExpenses(t *testing.T) {
	query := v1AnalysisQuery{
		StartDate: "2026-01-01", EndDate: "2026-02-28", GroupBy: "month",
		Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
	}
	transactions := []model.V1AnalyticsTransaction{
		{TransactionId: "1", TransactionDate: "2026-01-10", SignedAmount: -1000, FixedFlg: true, CategoryId: "22", CategoryName: "住宅"},
		{TransactionId: "2", TransactionDate: "2026-02-10", SignedAmount: -500, FixedFlg: true, CategoryId: "21", CategoryName: "通信"},
		{TransactionId: "3", TransactionDate: "2026-02-11", SignedAmount: -500, FixedFlg: false, CategoryId: "1", CategoryName: "食費"},
		{TransactionId: "4", TransactionDate: "2026-02-12", SignedAmount: 2000, FixedFlg: true, CategoryId: "27", CategoryName: "給与"},
	}
	response, err := buildV1FixedResponse(query, transactions)
	if err != nil {
		t.Fatal(err)
	}
	if response.Summary.ExpenseAmount != 1500 || response.Summary.MonthlyAverage != 750 || response.Summary.AnnualizedAmount != 9000 {
		t.Fatalf("unexpected summary: %+v", response.Summary)
	}
	if response.Summary.TotalExpenseRatio != 75 {
		t.Fatalf("total expense ratio = %v", response.Summary.TotalExpenseRatio)
	}
	if len(response.CategoryList) != 2 || len(response.Series) != 2 {
		t.Fatalf("unexpected response: %+v", response)
	}
}
