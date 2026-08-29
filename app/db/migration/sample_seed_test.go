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
	paymentIDs := map[string]uint64{"楽天カード": 1, "現金": 2, "PayPay": 3}

	rows := buildSampleTransactions(
		2,
		time.Date(2026, time.August, 29, 12, 0, 0, 0, location),
		masterIDs,
		customIDs,
		paymentIDs,
	)
	if len(rows) != 72 {
		t.Fatalf("transaction count = %d, want 72", len(rows))
	}

	monthCounts := map[string]int{}
	paymentCounts := map[uint64]int{}
	for _, row := range rows {
		monthCounts[row.TransactionDate[:7]]++
		if row.TransactionTime == nil || *row.TransactionTime == "" {
			t.Errorf("transaction %q has no transaction_time", row.TransactionName)
		}
		if row.PaymentID == nil {
			t.Errorf("transaction %q has no payment resource", row.TransactionName)
		} else {
			paymentCounts[*row.PaymentID]++
		}
	}
	for _, month := range []string{"2026-03", "2026-04", "2026-05", "2026-06", "2026-07", "2026-08"} {
		if monthCounts[month] != 12 {
			t.Errorf("month %s count = %d, want 12", month, monthCounts[month])
		}
	}
	for paymentID := uint64(1); paymentID <= 3; paymentID++ {
		if paymentCounts[paymentID] == 0 {
			t.Errorf("payment resource %d is unused", paymentID)
		}
	}
}
