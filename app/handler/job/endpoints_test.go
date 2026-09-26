package job

import (
	"MoneyHook/MoneyHook-API/model"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type jobStoreStub struct {
	rows     []model.JobMonthlyTransaction
	readErr  error
	writeErr error
	inserted []model.JobTransaction
}

func (s *jobStoreStub) SelectMonthlyTransaction(int, bool) (*[]model.JobMonthlyTransaction, error) {
	return &s.rows, s.readErr
}

func (s *jobStoreStub) InsertTransaction(rows *[]model.JobTransaction) error {
	s.inserted = *rows
	return s.writeErr
}

func TestProcessDailyJob(t *testing.T) {
	for _, test := range []struct {
		name        string
		store       jobStoreStub
		status      int
		body        string
		insertCount int
	}{
		{name: "read failure", store: jobStoreStub{readErr: errors.New("database unavailable")}, status: 500},
		{name: "no rows", status: 200, body: "Today is Nothing, Success Jobs"},
		{name: "insert rows", store: jobStoreStub{rows: []model.JobMonthlyTransaction{{UserNo: "2", MonthlyTransactionName: "Rent", MonthlyTransactionAmount: -1000}}}, status: 200, body: "Success Jobs", insertCount: 1},
		{name: "insert failure", store: jobStoreStub{rows: []model.JobMonthlyTransaction{{UserNo: "2"}}, writeErr: errors.New("insert failed")}, status: 422, insertCount: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/job/daily", nil)
			request.Header.Set(model.UserAgent, "Google-Cloud-Scheduler")
			request.Header.Set(model.ContentType, "application/octet-stream")
			request.Header.Set(model.XCloudScheduler, "true")
			request.Header.Set(model.XCloudSchedulerJobName, "daily")
			request.Header.Set(model.XCloudSchedulerScheduleTime, "2026-09-18T00:00:00Z")
			recorder := httptest.NewRecorder()
			if err := New(&test.store, "daily").ProcessDailyJob(echo.New().NewContext(request, recorder)); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != test.status || (test.body != "" && recorder.Body.String() != test.body) {
				t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
			}
			if len(test.store.inserted) != test.insertCount {
				t.Fatalf("inserted = %+v", test.store.inserted)
			}
			if test.store.readErr != nil {
				var body map[string]any
				if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body["status"] != "error" || body["code"] != "INTERNAL_ERROR" || body["message"] != "Failed to select monthly transactions" {
					t.Fatalf("unexpected error contract: %v", body)
				}
			}
		})
	}
}

func TestProcessDailyJobRejectsInvalidSchedulerHeaders(t *testing.T) {
	for _, test := range []struct {
		name   string
		header string
		value  string
	}{
		{name: "user agent", header: model.UserAgent, value: "browser"},
		{name: "content type", header: model.ContentType, value: "application/json"},
		{name: "scheduler marker", header: model.XCloudScheduler, value: "false"},
		{name: "job name", header: model.XCloudSchedulerJobName, value: "other"},
		{name: "schedule time", header: model.XCloudSchedulerScheduleTime, value: ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &jobStoreStub{}
			request := httptest.NewRequest(http.MethodPost, "/api/job/daily", nil)
			request.Header.Set(model.UserAgent, "Google-Cloud-Scheduler")
			request.Header.Set(model.ContentType, "application/octet-stream")
			request.Header.Set(model.XCloudScheduler, "true")
			request.Header.Set(model.XCloudSchedulerJobName, "daily")
			request.Header.Set(model.XCloudSchedulerScheduleTime, "2026-09-18T00:00:00Z")
			request.Header.Set(test.header, test.value)
			recorder := httptest.NewRecorder()

			if err := New(store, "daily").ProcessDailyJob(echo.New().NewContext(request, recorder)); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != http.StatusForbidden {
				t.Fatalf("status = %d", recorder.Code)
			}
			if len(store.inserted) != 0 {
				t.Fatalf("inserted = %+v", store.inserted)
			}
		})
	}
}
