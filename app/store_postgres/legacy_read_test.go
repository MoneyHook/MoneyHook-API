package store_postgres

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

func TestLegacyReadStoresPropagateDatabaseErrors(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "user=test"}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.Close()
	store := NewTransactionStore(db)
	for name, read := range map[string]func() error{
		"GetTimelineData":            func() error { _, err := store.GetTimelineData("2", "2026-09-01"); return err },
		"GetMonthlySpendingData":     func() error { _, err := store.GetMonthlySpendingData("2", "2026-09-01"); return err },
		"GetTransactionData":         func() error { _, err := store.GetTransactionData("2", "7"); return err },
		"GetMonthlyFixedData":        func() error { _, err := store.GetMonthlyFixedData("2", "2026-09-01", true); return err },
		"GetHome":                    func() error { _, err := store.GetHome("2", "2026-09-01"); return err },
		"GetMonthlyVariableData":     func() error { _, err := store.GetMonthlyVariableData("2", "2026-09-01"); return err },
		"GetTotalSpending":           func() error { _, err := store.GetTotalSpending("2", "", "", "2026-09-01", "2026-09-01"); return err },
		"GetGroupByPayment":          func() error { _, err := store.GetGroupByPayment("2", "2026-09-01"); return err },
		"GetLastMonthGroupByPayment": func() error { _, err := store.GetLastMonthGroupByPayment("2", "2026-09-01"); return err },
		"GetMonthlyWithdrawalAmount": func() error {
			_, err := store.GetMonthlyWithdrawalAmount("2", "1", "2026-09-01", "2026-09-30")
			return err
		},
		"GetFrequentTransactionName": func() error { _, err := store.GetFrequentTransactionName("2", 20); return err },
		"GetPaymentResourceList":     func() error { _, err := NewPaymentResourceStore(db).GetPaymentResourceList("2"); return err },
	} {
		t.Run(name, func(t *testing.T) {
			if err := read(); err == nil {
				t.Fatal("database failure was discarded")
			}
		})
	}
}
