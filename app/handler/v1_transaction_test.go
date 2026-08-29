package handler

import "testing"

func TestValidateV1TransactionInput(t *testing.T) {
	paymentID := "3"
	timeValue := "12:30"
	fixedFlg := false
	valid := v1TransactionInput{
		TransactionDate: "2026-08-29",
		TransactionTime: &timeValue,
		TransactionName: "ランチ",
		Amount:          1200,
		Sign:            -1,
		CategoryId:      "2",
		SubCategoryId:   "4",
		FixedFlg:        &fixedFlg,
		PaymentId:       &paymentID,
	}
	if errors := validateV1TransactionInput(valid); len(errors) != 0 {
		t.Fatalf("valid input returned errors: %v", errors)
	}

	invalidTime := "24:00"
	invalidPaymentID := "0"
	invalid := v1TransactionInput{
		TransactionDate: "2026-02-30",
		TransactionTime: &invalidTime,
		TransactionName: "",
		Amount:          0,
		Sign:            2,
		CategoryId:      "x",
		SubCategoryId:   "",
		PaymentId:       &invalidPaymentID,
	}
	errors := validateV1TransactionInput(invalid)
	for _, field := range []string{
		"transaction.transaction_date",
		"transaction.transaction_time",
		"transaction.transaction_name",
		"transaction.amount",
		"transaction.sign",
		"transaction.category_id",
		"transaction.sub_category_id",
		"transaction.fixed_flg",
		"transaction.payment_id",
	} {
		if _, exists := errors[field]; !exists {
			t.Errorf("missing validation error for %s: %v", field, errors)
		}
	}
}

func TestIsPositiveNumericID(t *testing.T) {
	for _, value := range []string{"1", "999999999999"} {
		if !isPositiveNumericID(value) {
			t.Errorf("%q should be valid", value)
		}
	}
	for _, value := range []string{"", "0", "-1", "1.5", "abc"} {
		if isPositiveNumericID(value) {
			t.Errorf("%q should be invalid", value)
		}
	}
}
