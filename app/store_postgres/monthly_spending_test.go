package store_postgres

import (
	"MoneyHook/MoneyHook-API/model"
	"reflect"
	"testing"
)

func TestFillMonthlySpendingData(t *testing.T) {
	rows := []model.MonthlySpendingData{{Month: "2026-01-01", TotalAmount: -100}, {Month: "2025-11-01", TotalAmount: -300}}
	got, err := fillMonthlySpendingData(rows, "2026-02-01")
	want := []model.MonthlySpendingData{
		{Month: "2026-02-01"}, {Month: "2026-01-01", TotalAmount: -100},
		{Month: "2025-12-01"}, {Month: "2025-11-01", TotalAmount: -300},
		{Month: "2025-10-01"}, {Month: "2025-09-01"},
	}
	if err != nil || !reflect.DeepEqual(*got, want) {
		t.Fatalf("got %v, %v; want %v", got, err, want)
	}
	if _, err := fillMonthlySpendingData(nil, "invalid"); err == nil {
		t.Fatal("invalid date was accepted")
	}
}
