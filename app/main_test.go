package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestHealthRouteIsPublic(t *testing.T) {
	e := echo.New()
	registerHealthRoute(e)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()
	e.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.String() != "Success, running" {
		t.Fatalf("body = %q, want %q", recorder.Body.String(), "Success, running")
	}
}
