package household

import (
	h "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/router"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type pagedEntries struct {
	h.Store
	filters []h.Filter
}

func (s *pagedEntries) Entries(_ context.Context, _, _ string, filter h.Filter) (*h.EntryPage, error) {
	s.filters = append(s.filters, filter)
	member := "2"
	if filter.Cursor == "" {
		next := "100"
		return &h.EntryPage{Entries: []h.Entry{
			{ID: "1", TransactionInput: h.TransactionInput{Amount: 100, Sign: -1, CategoryID: "1"}, CategoryName: "食費", Payer: h.Payer{Kind: "common", DisplayName: "家族共通"}},
			{ID: "2", TransactionInput: h.TransactionInput{Amount: 900, Sign: -1, CategoryID: "1"}, Excluded: true},
		}, NextCursor: &next}, nil
	}
	return &h.EntryPage{Entries: []h.Entry{
		{ID: "3", TransactionInput: h.TransactionInput{Amount: 200, Sign: -1, CategoryID: "1"}, CategoryName: "食費", Payer: h.Payer{Kind: "member", MemberID: &member, DisplayName: "本人"}},
		{ID: "4", TransactionInput: h.TransactionInput{Amount: 1000, Sign: 1, CategoryID: "2"}},
	}}, nil
}

func TestAnalyticsAggregatesAllPagesAndExcludesCorrections(t *testing.T) {
	for _, group := range []string{"categories", "payers"} {
		t.Run(group, func(t *testing.T) {
			store := &pagedEntries{}
			recorder := httptest.NewRecorder()
			c := echo.New().NewContext(httptest.NewRequest("GET", "/?month=2026-09-01", nil), recorder)
			c.Set(router.ContextKeyUserNo, "2")
			c.SetParamNames("householdId", "group")
			c.SetParamValues("1", group)
			if err := New(store).Analytics(c); err != nil {
				t.Fatal(err)
			}
			var result h.Summary
			if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != 200 || result.Income != 1000 || result.Expense != 300 || result.Balance != 700 {
				t.Fatalf("bad totals: %s", recorder.Body.String())
			}
			var grouped int64
			for _, g := range result.Groups {
				grouped += g.Amount
			}
			if grouped != result.Expense {
				t.Fatal("group total differs from expense")
			}
			if len(store.filters) != 2 || store.filters[1].Cursor != "100" || store.filters[1].Month != "2026-09-01" {
				t.Fatal(store.filters)
			}
			if group == "categories" && (len(result.Groups) != 1 || result.Groups[0].ID != "1") {
				t.Fatal(result.Groups)
			}
			if group == "payers" && (len(result.Groups) != 2 || result.Groups[0].ID != "2" || result.Groups[1].ID != "common") {
				t.Fatal(result.Groups)
			}
		})
	}
}
