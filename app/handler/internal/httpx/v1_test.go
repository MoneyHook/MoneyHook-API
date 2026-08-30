package httpx

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestDecodeV1JSONRejectsUnknownAndTrailingValues(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "unknown field", body: `{"transaction":{"unknown":true}}`},
		{name: "trailing value", body: `{"transaction":{}} {}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			request := httptest.NewRequest("POST", "/api/v1/transactions", strings.NewReader(test.body))
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			context := e.NewContext(request, httptest.NewRecorder())
			var destination struct {
				Transaction struct{} `json:"transaction"`
			}

			if err := DecodeV1JSON(context, &destination); err == nil {
				t.Fatal("DecodeV1JSON accepted an invalid request body")
			}
		})
	}
}
