package household

import (
	h "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/router"
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"net/http/httptest"
	"reflect"
	"testing"
)

type analyticsStore struct {
	h.Store
	month, group string
	filter       h.DuplicateFilter
	calls        int
}

func (s *analyticsStore) Analytics(_ context.Context, u, f, month, group string) (*h.Summary, error) {
	if u != "2" || f != "1" {
		return nil, h.NotFound
	}
	s.month, s.group = month, group
	s.calls++
	return &h.Summary{Income: 1000, Expense: 300, Balance: 700, Groups: []h.SummaryGroup{{ID: "1", Name: "食費", Amount: 300}}}, nil
}
func (s *analyticsStore) Duplicates(_ context.Context, u, f string, filter h.DuplicateFilter) ([]h.Entry, error) {
	s.filter = filter
	s.calls++
	return []h.Entry{}, nil
}
func TestAnalyticsUsesStoreAggregation(t *testing.T) {
	for _, group := range []string{"overview", "categories", "payers"} {
		s := &analyticsStore{}
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(httptest.NewRequest("GET", "/?month=2026-09-01", nil), rec)
		c.Set(router.ContextKeyUserNo, "2")
		c.SetParamNames("householdId", "group")
		c.SetParamValues("1", group)
		if err := New(s).Analytics(c); err != nil {
			t.Fatal(err)
		}
		var result h.Summary
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if rec.Code != 200 || result.Income != 1000 || result.Expense != 300 || result.Balance != 700 || s.calls != 1 || s.group != group || s.month != "2026-09-01" {
			t.Fatalf("bad response: %s", rec.Body.String())
		}
	}
}
func TestDuplicatesPassesValidatedConditions(t *testing.T) {
	for _, query := range []string{
		"date=2026-09-10&amount=100&sign=-1&payer=common&exclude=10",
		"date=invalid&amount=100&sign=-1", "date=2026-09-10&amount=0&sign=-1", "date=2026-09-10&amount=100&sign=0", "date=2026-09-10&amount=100&sign=-1&exclude=bad",
	} {
		s := &analyticsStore{}
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(httptest.NewRequest("GET", "/?"+query, nil), rec)
		c.Set(router.ContextKeyUserNo, "2")
		c.SetParamNames("householdId")
		c.SetParamValues("1")
		if err := New(s).Duplicates(c); err != nil {
			t.Fatal(err)
		}
		if query == "date=2026-09-10&amount=100&sign=-1&payer=common&exclude=10" {
			want := h.DuplicateFilter{Date: "2026-09-10", Amount: 100, Sign: -1, Payer: "common", Exclude: "10"}
			if rec.Code != 200 || s.calls != 1 || !reflect.DeepEqual(s.filter, want) {
				t.Fatalf("bad call: %+v, %s", s, rec.Body.String())
			}
		} else if rec.Code != 422 || s.calls != 0 {
			t.Fatalf("invalid request reached store: %s", rec.Body.String())
		}
	}
}
