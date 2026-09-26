package handler

import (
	jobdomain "MoneyHook/MoneyHook-API/job"
	"MoneyHook/MoneyHook-API/model"
	"MoneyHook/MoneyHook-API/router"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"google.golang.org/api/idtoken"
)

type routeSchedulerTokenValidator struct{}

func (routeSchedulerTokenValidator) Validate(context.Context, string, string) (*idtoken.Payload, error) {
	return &idtoken.Payload{
		Issuer: "https://accounts.google.com",
		Claims: map[string]interface{}{
			"email":          "scheduler@example-project.iam.gserviceaccount.com",
			"email_verified": true,
		},
	}, nil
}

type routeJobStore struct{}

var _ jobdomain.Store = routeJobStore{}

func (routeJobStore) SelectMonthlyTransaction(int, bool) (*[]model.JobMonthlyTransaction, error) {
	rows := []model.JobMonthlyTransaction{}
	return &rows, nil
}

func (routeJobStore) InsertTransaction(*[]model.JobTransaction) error { return nil }

func routeSchedulerConfig() router.SchedulerAuthConfig {
	return router.SchedulerAuthConfig{
		Audience:            "https://api.example.com/api/job/daily",
		ServiceAccountEmail: "scheduler@example-project.iam.gserviceaccount.com",
		JobName:             "daily",
	}
}

var expectedBusinessRoutes = map[string]struct{}{
	"GET /api/v1/budget":                                       {},
	"PUT /api/v1/budget":                                       {},
	"GET /api/v1/settings":                                     {},
	"PATCH /api/v1/settings":                                   {},
	"POST /api/v1/transactions":                                {},
	"GET /api/v1/transactions/:transactionId":                  {},
	"PATCH /api/v1/transactions/:transactionId":                {},
	"DELETE /api/v1/transactions/:transactionId":               {},
	"GET /api/v1/analytics/overview":                           {},
	"GET /api/v1/analytics/categories":                         {},
	"GET /api/v1/analytics/fixed":                              {},
	"GET /api/v1/analytics/payments":                           {},
	"GET /api/transaction/getTimelineData":                     {},
	"GET /api/transaction/getMonthlySpendingData":              {},
	"GET /api/transaction/getTransaction/:transactionId":       {},
	"GET /api/transaction/getMonthlyFixedIncome":               {},
	"GET /api/transaction/getMonthlyFixedSpending":             {},
	"GET /api/transaction/getHome":                             {},
	"GET /api/transaction/getMonthlyVariableData":              {},
	"GET /api/transaction/getTotalSpending":                    {},
	"GET /api/transaction/groupByPayment":                      {},
	"GET /api/transaction/getMonthlyWithdrawalAmount":          {},
	"GET /api/transaction/getFrequentTransactionName":          {},
	"POST /api/transaction/addTransaction":                     {},
	"POST /api/transaction/addTransactionList":                 {},
	"PATCH /api/transaction/editTransaction":                   {},
	"DELETE /api/transaction/deleteTransaction/:transactionId": {},
	"GET /api/fixed/getFixed":                                  {},
	"GET /api/fixed/getDeletedFixed":                           {},
	"POST /api/fixed/addFixed":                                 {},
	"PATCH /api/fixed/editFixed":                               {},
	"DELETE /api/fixed/deleteFixed/:monthly_transaction_id":    {},
	"GET /api/category/getCategoryList":                        {},
	"GET /api/category/getCategoryWithSubCategoryList":         {},
	"GET /api/subCategory/getSubCategoryList/:categoryId":      {},
	"POST /api/subCategory/editSubCategory":                    {},
	"GET /api/payment/getPayment":                              {},
	"POST /api/payment/addPayment":                             {},
	"PATCH /api/payment/editPayment":                           {},
	"PUT /api/payment/reorder":                                 {},
	"DELETE /api/payment/deletePayment/:paymentId":             {},
	"GET /api/payment/getPaymentType":                          {},
	"POST /api/job/daily":                                      {},
}

func TestRegisterIncludesCompleteBusinessRouteSet(t *testing.T) {
	e := echo.New()
	New(Dependencies{}).Register(e.Group("/api"))

	actual := map[string]struct{}{}
	for _, route := range e.Routes() {
		if route.Method == "echo_route_not_found" {
			continue
		}
		actual[route.Method+" "+route.Path] = struct{}{}
	}

	if len(actual) != len(expectedBusinessRoutes) {
		t.Fatalf("business route count = %d, want %d: %v", len(actual), len(expectedBusinessRoutes), actual)
	}
	for route := range expectedBusinessRoutes {
		if _, exists := actual[route]; !exists {
			t.Errorf("missing route %s", route)
		}
	}
}

func TestUserBusinessRoutesRequireFirebaseBearerAuthentication(t *testing.T) {
	e := echo.New()
	New(Dependencies{}).Register(e.Group("/api"))

	for route := range expectedBusinessRoutes {
		if route == "POST /api/job/daily" {
			continue
		}
		method, path, _ := strings.Cut(route, " ")
		path = strings.ReplaceAll(path, ":transactionId", "1")
		path = strings.ReplaceAll(path, ":monthly_transaction_id", "1")
		path = strings.ReplaceAll(path, ":categoryId", "1")
		path = strings.ReplaceAll(path, ":paymentId", "1")

		t.Run(route, func(t *testing.T) {
			request := httptest.NewRequest(method, path, nil)
			recorder := httptest.NewRecorder()
			e.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
			}
			var response struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode authentication response: %v", err)
			}
			if response.Code != "UNAUTHORIZED" {
				t.Fatalf("code = %q, want UNAUTHORIZED", response.Code)
			}
		})
	}
}

func TestDailyJobAcceptsOnlyConfiguredSchedulerIdentity(t *testing.T) {
	e := echo.New()
	New(Dependencies{
		SchedulerToken: routeSchedulerTokenValidator{},
		SchedulerAuth:  routeSchedulerConfig(),
		JobStore:       routeJobStore{},
	}).Register(e.Group("/api"))

	request := httptest.NewRequest(http.MethodPost, "/api/job/daily", nil)
	request.Header.Set(echo.HeaderAuthorization, "Bearer scheduler-token")
	request.Header.Set(model.UserAgent, "Google-Cloud-Scheduler")
	request.Header.Set(model.ContentType, "application/octet-stream")
	request.Header.Set(model.XCloudScheduler, "true")
	request.Header.Set(model.XCloudSchedulerJobName, "daily")
	request.Header.Set(model.XCloudSchedulerScheduleTime, "2026-09-18T00:00:00Z")
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || recorder.Body.String() != "Today is Nothing, Success Jobs" {
		t.Fatalf("response = %d %q", recorder.Code, recorder.Body.String())
	}
}
