//go:build integration

package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"errors"
	"testing"
)

func TestHouseholdAnalyticsQueriesIntegration(t *testing.T) {
	_, store, counted, counter, _, _, _ := householdReadFixture(t)
	ctx := context.Background()
	t.Run("aggregate all rows and preserve group names", func(t *testing.T) {
		for _, group := range []string{"overview", "categories", "payers"} {
			counter.count = 0
			summary, err := counted.Analytics(ctx, "2", "1", "2026-09-01", group)
			if err != nil {
				t.Fatal(err)
			}
			if counter.count != 3 || summary.Income != 1300 || summary.Expense != 100700 || summary.Balance != -99400 {
				t.Fatalf("%s: queries=%d result=%+v", group, counter.count, summary)
			}
			groups := map[string]h.SummaryGroup{}
			for _, g := range summary.Groups {
				groups[g.ID] = g
			}
			if group == "payers" {
				if len(groups) != 3 || groups["2"].Name != "保存時の名前" || groups["2"].Amount != 400 || groups["common"].Amount != 200 {
					t.Fatal(groups)
				}
			} else if len(groups) != 2 || groups["1"].Amount != 100100 || groups["2"].Amount != 600 {
				t.Fatal(groups)
			}
		}
		summary, err := store.Analytics(ctx, "2", "1", "2026-10-01", "categories")
		if err != nil || summary.Expense != 850 {
			t.Fatalf("moved correction: %+v %v", summary, err)
		}
		summary, err = store.Analytics(ctx, "2", "1", "2026-11-01", "overview")
		if err != nil || summary.Income != 0 || len(summary.Groups) != 0 {
			t.Fatalf("empty month: %+v %v", summary, err)
		}
	})
	t.Run("duplicates filter in SQL without page truncation", func(t *testing.T) {
		for _, tc := range []struct {
			filter h.DuplicateFilter
			want   int
		}{
			{h.DuplicateFilter{Date: "2026-09-10", Amount: 100, Sign: -1}, 1001},
			{h.DuplicateFilter{Date: "2026-09-10", Amount: 100, Sign: -1, Exclude: "1", Payer: "1"}, 1000},
			{h.DuplicateFilter{Date: "2026-09-10", Amount: 100, Sign: -1, Payer: "common"}, 0},
			{h.DuplicateFilter{Date: "2026-09-11", Amount: 200, Sign: -1, Payer: "common"}, 1},
			{h.DuplicateFilter{Date: "2026-09-12", Amount: 400, Sign: -1}, 1},
			{h.DuplicateFilter{Date: "2026-09-12", Amount: 9999, Sign: -1}, 0},
			{h.DuplicateFilter{Date: "2026-10-01", Amount: 850, Sign: -1}, 1},
			{h.DuplicateFilter{Date: "2026-09-11", Amount: 300, Sign: 1}, 1},
		} {
			counter.count = 0
			rows, err := counted.Duplicates(ctx, "2", "1", tc.filter)
			if err != nil {
				t.Fatal(err)
			}
			if counter.count != 3 || len(rows) != tc.want || rows == nil {
				t.Fatalf("%+v: rows=%d queries=%d", tc.filter, len(rows), counter.count)
			}
		}
	})
	t.Run("access and validation", func(t *testing.T) {
		if _, err := store.Entries(ctx, "5", "1", h.Filter{Month: "2026-09-01"}); !errors.Is(err, h.NotFound) {
			t.Fatal(err)
		}
		if _, err := store.Analytics(ctx, "2", "2", "2026-09-01", "overview"); !errors.Is(err, h.NotFound) {
			t.Fatal(err)
		}
		if _, err := store.Duplicates(ctx, "5", "1", h.DuplicateFilter{Date: "2026-09-10", Amount: 100, Sign: -1}); !errors.Is(err, h.NotFound) {
			t.Fatal(err)
		}
		if _, err := store.Analytics(ctx, "2", "1", "2026-09-02", "overview"); !errors.Is(err, h.Invalid) {
			t.Fatal(err)
		}
		if _, err := store.Duplicates(ctx, "2", "1", h.DuplicateFilter{Date: "2026-09-10", Amount: 100, Sign: -1, Payer: "bad"}); !errors.Is(err, h.Invalid) {
			t.Fatal(err)
		}
	})
}
