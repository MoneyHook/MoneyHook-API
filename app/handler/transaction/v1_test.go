package transaction

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
	validCreate := v1TransactionCreateInput{
		TransactionDate: valid.TransactionDate,
		TransactionTime: valid.TransactionTime,
		TransactionName: valid.TransactionName,
		Amount:          valid.Amount,
		Sign:            valid.Sign,
		CategoryId:      valid.CategoryId,
		SubCategoryId:   valid.SubCategoryId,
		FixedFlg:        valid.FixedFlg,
		PaymentId:       valid.PaymentId,
	}
	if errors := validateV1TransactionCreateInput(validCreate); len(errors) != 0 {
		t.Fatalf("valid input returned errors: %v", errors)
	}
	if errors := validateV1TransactionUpdateInput(valid); len(errors) != 0 {
		t.Fatalf("valid update input returned errors: %v", errors)
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
	errors := validateV1TransactionUpdateInput(invalid)
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

func TestValidateV1TransactionCreateSubCategoryChoice(t *testing.T) {
	fixedFlg := false
	base := v1TransactionInput{
		TransactionDate: "2026-09-25",
		TransactionName: "ランチ",
		Amount:          1200,
		Sign:            -1,
		CategoryId:      "2",
		FixedFlg:        &fixedFlg,
	}

	withName := v1TransactionCreateInput{
		TransactionDate: base.TransactionDate,
		TransactionName: base.TransactionName,
		Amount:          base.Amount,
		Sign:            base.Sign,
		CategoryId:      base.CategoryId,
		SubCategoryName: "  新しい分類  ",
		FixedFlg:        base.FixedFlg,
	}
	if errors := validateV1TransactionCreateInput(withName); len(errors) != 0 {
		t.Fatalf("name-only input returned errors: %v", errors)
	}

	withBoth := withName
	withBoth.SubCategoryId = "4"
	if _, exists := validateV1TransactionCreateInput(withBoth)["transaction.sub_category_id"]; !exists {
		t.Fatal("ID and name together must be rejected")
	}

	withNeither := withName
	withNeither.SubCategoryName = ""
	if _, exists := validateV1TransactionCreateInput(withNeither)["transaction.sub_category_id"]; !exists {
		t.Fatal("missing ID and name must be rejected")
	}

	tooLong := withName
	tooLong.SubCategoryName = "あいうえおかきくけこさしすせそたち"
	if _, exists := validateV1TransactionCreateInput(tooLong)["transaction.sub_category_name"]; !exists {
		t.Fatal("a name longer than 16 characters must be rejected")
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
