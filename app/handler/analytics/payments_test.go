package analytics

import (
	"MoneyHook/MoneyHook-API/model"
	"testing"
	"time"
)

func TestBuildV1PaymentsResponseIncludesUnclassifiedExpenses(t *testing.T) {
	query := v1AnalysisQuery{
		StartDate: "2026-01-01", EndDate: "2026-02-28", GroupBy: "month",
		Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC),
	}
	paymentID := "1"
	paymentName := "現金"
	transactions := []model.V1AnalyticsTransaction{
		{TransactionId: "1", TransactionDate: "2026-01-10", SignedAmount: -1000, PaymentId: &paymentID, PaymentName: &paymentName},
		{TransactionId: "2", TransactionDate: "2026-02-10", SignedAmount: -500, PaymentId: &paymentID, PaymentName: &paymentName},
		{TransactionId: "3", TransactionDate: "2026-02-11", SignedAmount: -300},
		{TransactionId: "4", TransactionDate: "2026-02-12", SignedAmount: 2000, PaymentId: &paymentID, PaymentName: &paymentName},
	}
	response, err := buildV1PaymentsResponse(query, transactions)
	if err != nil {
		t.Fatal(err)
	}
	if response.TotalExpenseAmount != 1800 || len(response.PaymentList) != 2 {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.PaymentList[0].PaymentName != "現金" || response.PaymentList[0].TransactionCount != 2 || response.PaymentList[0].AverageAmount != 750 {
		t.Fatalf("unexpected classified payment: %+v", response.PaymentList[0])
	}
	if response.PaymentList[1].PaymentId != nil || response.PaymentList[1].PaymentName != "未分類" {
		t.Fatalf("unexpected unclassified payment: %+v", response.PaymentList[1])
	}
}
