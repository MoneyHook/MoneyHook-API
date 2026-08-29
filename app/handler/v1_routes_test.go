package handler

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRegisterIncludesCompleteV1RouteSet(t *testing.T) {
	e := echo.New()
	(&Handler{}).Register(e.Group("/api"))

	want := map[string]bool{
		http.MethodPost + " /api/v1/transactions":                  true,
		http.MethodGet + " /api/v1/transactions/:transactionId":    true,
		http.MethodPatch + " /api/v1/transactions/:transactionId":  true,
		http.MethodDelete + " /api/v1/transactions/:transactionId": true,
		http.MethodGet + " /api/v1/analytics/overview":             true,
		http.MethodGet + " /api/v1/analytics/categories":           true,
		http.MethodGet + " /api/v1/analytics/fixed":                true,
		http.MethodGet + " /api/v1/analytics/payments":             true,
	}
	got := map[string]bool{}
	for _, route := range e.Routes() {
		if route.Method == "echo_route_not_found" {
			continue
		}
		if len(route.Path) >= len("/api/v1/") && route.Path[:len("/api/v1/")] == "/api/v1/" {
			got[route.Method+" "+route.Path] = true
		}
	}

	if len(got) != len(want) {
		t.Fatalf("v1 route count = %d, want %d: %v", len(got), len(want), got)
	}
	for route := range want {
		if !got[route] {
			t.Errorf("missing route %s", route)
		}
	}
}
