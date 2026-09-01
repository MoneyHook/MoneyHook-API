package analytics

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestParseV1AnalysisQuery(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantError string
	}{
		{name: "366 day leap-year range", query: "start_date=2024-01-01&end_date=2024-12-31"},
		{name: "too long", query: "start_date=2024-01-01&end_date=2025-01-01", wantError: "date_range"},
		{name: "reversed", query: "start_date=2026-08-30&end_date=2026-08-29", wantError: "date_range"},
		{name: "invalid date", query: "start_date=2026-02-30&end_date=2026-03-01", wantError: "start_date"},
		{name: "invalid grouping", query: "start_date=2026-01-01&end_date=2026-01-31&group_by=year", wantError: "group_by"},
		{name: "invalid comparison", query: "start_date=2026-01-01&end_date=2026-01-31&compare=current", wantError: "compare"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			request := httptest.NewRequest("GET", "/api/v1/analytics/overview?"+test.query, nil)
			query, errors := parseV1AnalysisQuery(e.NewContext(request, httptest.NewRecorder()))
			if test.wantError == "" {
				if len(errors) != 0 {
					t.Fatalf("unexpected errors: %v", errors)
				}
				if query.GroupBy != "month" || query.Compare != "none" {
					t.Fatalf("defaults = group_by:%s compare:%s", query.GroupBy, query.Compare)
				}
				return
			}
			if _, ok := errors[test.wantError]; !ok {
				t.Fatalf("missing %s error: %v", test.wantError, errors)
			}
		})
	}
}

func TestPreviousPeriodUsesSameInclusiveDayCount(t *testing.T) {
	e := echo.New()
	request := httptest.NewRequest("GET", "/api/v1/analytics/overview?start_date=2026-03-01&end_date=2026-03-31&compare=previous_period", nil)
	query, errors := parseV1AnalysisQuery(e.NewContext(request, httptest.NewRecorder()))
	if len(errors) != 0 {
		t.Fatalf("unexpected errors: %v", errors)
	}
	start, end := query.previousPeriod()
	if start != "2026-01-29" || end != "2026-02-28" {
		t.Fatalf("previous period = %s..%s", start, end)
	}
}
