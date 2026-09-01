package analytics

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

type v1AnalysisQuery struct {
	StartDate string
	EndDate   string
	GroupBy   string
	Compare   string
	Start     time.Time
	End       time.Time
}

func parseV1AnalysisQuery(c echo.Context) (v1AnalysisQuery, map[string]string) {
	query := v1AnalysisQuery{
		StartDate: c.QueryParam("start_date"),
		EndDate:   c.QueryParam("end_date"),
		GroupBy:   c.QueryParam("group_by"),
		Compare:   c.QueryParam("compare"),
	}
	if query.GroupBy == "" {
		query.GroupBy = "month"
	}
	if query.Compare == "" {
		query.Compare = "none"
	}

	fieldErrors := map[string]string{}
	start, startErr := time.Parse("2006-01-02", query.StartDate)
	if startErr != nil || start.Format("2006-01-02") != query.StartDate {
		fieldErrors["start_date"] = "YYYY-MM-DD形式の実在する日付を指定してください"
	}
	end, endErr := time.Parse("2006-01-02", query.EndDate)
	if endErr != nil || end.Format("2006-01-02") != query.EndDate {
		fieldErrors["end_date"] = "YYYY-MM-DD形式の実在する日付を指定してください"
	}
	if startErr == nil && endErr == nil {
		query.Start = start
		query.End = end
		if start.After(end) {
			fieldErrors["date_range"] = "start_dateはend_date以前にしてください"
		} else if int(end.Sub(start).Hours()/24)+1 > 366 {
			fieldErrors["date_range"] = "分析期間は366日以内にしてください"
		}
	}
	if query.GroupBy != "day" && query.GroupBy != "week" && query.GroupBy != "month" {
		fieldErrors["group_by"] = "day、week、monthのいずれかを指定してください"
	}
	if query.Compare != "none" && query.Compare != "previous_period" {
		fieldErrors["compare"] = "noneまたはprevious_periodを指定してください"
	}
	return query, fieldErrors
}

func (query v1AnalysisQuery) previousPeriod() (string, string) {
	days := int(query.End.Sub(query.Start).Hours()/24) + 1
	previousEnd := query.Start.AddDate(0, 0, -1)
	previousStart := previousEnd.AddDate(0, 0, -(days - 1))
	return previousStart.Format("2006-01-02"), previousEnd.Format("2006-01-02")
}

func analysisBucketStart(value time.Time, groupBy string) time.Time {
	switch groupBy {
	case "day":
		return value
	case "week":
		offset := (int(value.Weekday()) + 6) % 7
		return value.AddDate(0, 0, -offset)
	default:
		return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
}

func nextAnalysisBucket(value time.Time, groupBy string) time.Time {
	switch groupBy {
	case "day":
		return value.AddDate(0, 0, 1)
	case "week":
		return value.AddDate(0, 0, 7)
	default:
		return value.AddDate(0, 1, 0)
	}
}

func parseStoredTransactionDate(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse stored transaction date %q: %w", value, err)
	}
	return parsed, nil
}
