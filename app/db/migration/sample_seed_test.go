package migration

import (
	"testing"
	"time"
)

func TestBuildSampleTransactions(t *testing.T) {
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	masterIDs := map[string]uint64{}
	for index, definition := range requiredSampleMasterSubCategories() {
		masterIDs[sampleSubCategoryKey(definition.CategoryID, definition.Name)] = uint64(index + 1)
	}
	customIDs := map[string]uint64{}
	for index, definition := range sampleCustomSubCategories {
		customIDs[sampleSubCategoryKey(definition.CategoryID, definition.Name)] = uint64(index + 100)
	}
	hiddenIDs := map[uint64]struct{}{}
	for _, definition := range sampleHiddenSubCategories {
		hiddenIDs[masterIDs[sampleSubCategoryKey(definition.CategoryID, definition.Name)]] = struct{}{}
	}
	customIDSet := map[uint64]struct{}{}
	for _, customID := range customIDs {
		customIDSet[customID] = struct{}{}
	}
	rows := buildSampleTransactions(2, time.Date(2026, time.August, 29, 12, 0, 0, 0, location), masterIDs, customIDs, map[string]uint64{"楽天カード": 1, "現金": 2, "PayPay": 3})
	if got, want := len(rows), 84; got != want {
		t.Fatalf("transaction count = %d, want %d", got, want)
	}

	monthCounts := map[string]int{}
	paymentCounts := map[uint64]int{}
	variations := map[string]bool{
		"transaction_time":         false,
		"missing_transaction_time": false,
		"payment_resource":         false,
		"missing_payment_resource": false,
		"income":                   false,
		"expense":                  false,
		"fixed":                    false,
		"variable":                 false,
		"custom_sub_category":      false,
		"hidden_sub_category":      false,
	}
	for _, row := range rows {
		monthCounts[row.TransactionDate[:7]]++
		variations["transaction_time"] = variations["transaction_time"] || row.TransactionTime != nil
		variations["missing_transaction_time"] = variations["missing_transaction_time"] || row.TransactionTime == nil
		variations["payment_resource"] = variations["payment_resource"] || row.PaymentID != nil
		variations["missing_payment_resource"] = variations["missing_payment_resource"] || row.PaymentID == nil
		variations["income"] = variations["income"] || row.TransactionAmount > 0
		variations["expense"] = variations["expense"] || row.TransactionAmount < 0
		variations["fixed"] = variations["fixed"] || row.FixedFlg
		variations["variable"] = variations["variable"] || !row.FixedFlg
		_, isCustom := customIDSet[row.SubCategoryID]
		variations["custom_sub_category"] = variations["custom_sub_category"] || isCustom
		_, isHidden := hiddenIDs[row.SubCategoryID]
		variations["hidden_sub_category"] = variations["hidden_sub_category"] || isHidden
		if row.PaymentID != nil {
			paymentCounts[*row.PaymentID]++
		}
	}
	for month, want := range map[string]int{"2026-03": 12, "2026-04": 14, "2026-05": 13, "2026-06": 15, "2026-07": 14, "2026-08": 16} {
		if got := monthCounts[month]; got != want {
			t.Errorf("month %s count = %d, want %d", month, got, want)
		}
	}
	for name, exists := range variations {
		if !exists {
			t.Errorf("sample data has no %s variation", name)
		}
	}
	for paymentID := uint64(1); paymentID <= 3; paymentID++ {
		if paymentCounts[paymentID] == 0 {
			t.Errorf("payment resource %d is unused", paymentID)
		}
	}
}

func TestBuildSampleBudgets(t *testing.T) {
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	rows := buildSampleBudgets(2, time.Date(2026, time.August, 29, 12, 0, 0, 0, location))
	want := []budgetSchema{
		{UserNo: 2, EffectiveFrom: "2026-03-01", MonthlyBudgetAmount: 240000},
		{UserNo: 2, EffectiveFrom: "2026-04-01", MonthlyBudgetAmount: 240000},
		{UserNo: 2, EffectiveFrom: "2026-05-01", MonthlyBudgetAmount: 245000},
		{UserNo: 2, EffectiveFrom: "2026-06-01", MonthlyBudgetAmount: 250000},
		{UserNo: 2, EffectiveFrom: "2026-07-01", MonthlyBudgetAmount: 250000},
		{UserNo: 2, EffectiveFrom: "2026-08-01", MonthlyBudgetAmount: 260000},
	}
	if len(rows) != len(want) {
		t.Fatalf("budget count = %d, want %d", len(rows), len(want))
	}
	for index := range want {
		if rows[index] != want[index] {
			t.Errorf("budget[%d] = %#v, want %#v", index, rows[index], want[index])
		}
	}
}
