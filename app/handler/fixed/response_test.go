package fixed

import (
	"MoneyHook/MoneyHook-API/model"
	"testing"
)

func TestGetFixedDeletedResponseIncludesFieldsNeededToReactivate(t *testing.T) {
	paymentID := "8"
	response := GetFixedDeletedResponse(&[]model.GetDeletedFixed{
		{
			MonthlyTransactionId:     "12",
			MonthlyTransactionName:   "ジム",
			MonthlyTransactionAmount: 5000,
			MonthlyTransactionSign:   -1,
			MonthlyTransactionDate:   10,
			CategoryId:               "15",
			CategoryName:             "健康",
			SubCategoryId:            "18",
			SubCategoryName:          "ジム・フィットネス",
			PaymentId:                paymentID,
		},
	})

	if len(*response) != 1 {
		t.Fatalf("response length = %d, want 1", len(*response))
	}
	item := (*response)[0]
	if item.MonthlyTransactionSign != -1 || item.CategoryId != "15" || item.SubCategoryId != "18" {
		t.Fatalf("reactivation fields missing: %+v", item)
	}
	if item.PaymentId == nil || *item.PaymentId != paymentID {
		t.Fatalf("payment id = %v, want %q", item.PaymentId, paymentID)
	}
}
