package household

import (
	h "MoneyHook/MoneyHook-API/household"
	t "MoneyHook/MoneyHook-API/transaction"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRespondErrorContract(tst *testing.T) {
	for _, tc := range []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{"missing original", fmt.Errorf("share: %w", t.ErrNotFound), 404, "NOT_FOUND"},
		{"invalid personal reference", t.ErrInvalidRelation, 422, "VALIDATION_ERROR"},
		{"conflict", h.Conflict, 409, "VERSION_CONFLICT"},
		{"forbidden", h.Forbidden, 403, "FORBIDDEN"},
		{"rate limited", h.Fail("RATE_LIMITED", "試行回数を超えました"), 429, "RATE_LIMITED"},
		{"unexpected", errors.New("database connection secret"), 500, "INTERNAL_ERROR"},
	} {
		tst.Run(tc.name, func(tst *testing.T) {
			recorder := httptest.NewRecorder()
			c := echo.New().NewContext(httptest.NewRequest("GET", "/", nil), recorder)
			if err := respond(c, tc.err); err != nil {
				tst.Fatal(err)
			}
			var body struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				tst.Fatal(err)
			}
			if recorder.Code != tc.status || body.Code != tc.code {
				tst.Fatalf("got %d %s, want %d %s", recorder.Code, recorder.Body.String(), tc.status, tc.code)
			}
			if tc.status == 429 && recorder.Header().Get("Retry-After") != "900" {
				tst.Fatal("missing retry delay")
			}
		})
	}
}
