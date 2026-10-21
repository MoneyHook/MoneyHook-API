package household

import (
	h "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/router"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type bulkStore struct {
	h.Store
	calls int
	share bool
	input h.BulkSharingInput
	err   error
}

func (s *bulkStore) BulkShare(_ context.Context, u, f string, in h.BulkSharingInput, share bool) (*h.BulkSharingResponse, error) {
	s.calls++
	s.input, s.share = in, share
	if u != "2" || f != "1" {
		return nil, h.Forbidden
	}
	return &h.BulkSharingResponse{Results: []h.BulkSharingResult{{TransactionID: "10", Outcome: "failed", Code: "NOT_FOUND", Message: "見つかりません"}}}, s.err
}
func TestBulkSharingHandler(t *testing.T) {
	for _, tc := range []struct {
		name, body         string
		unshare, anonymous bool
		err                error
		status, calls      int
	}{
		{name: "share partial result", body: `{"transaction_ids":["10"]}`, status: 200, calls: 1},
		{name: "unshare", body: `{"transaction_ids":["10"]}`, unshare: true, status: 200, calls: 1},
		{name: "empty", body: `{"transaction_ids":[]}`, status: 422},
		{name: "missing", body: `{}`, status: 422},
		{name: "invalid ID", body: `{"transaction_ids":["0"]}`, status: 422},
		{name: "unknown field", body: `{"transaction_ids":["10"],"other":1}`, status: 400},
		{name: "invalid JSON", body: `{`, status: 400},
		{name: "unauthenticated", body: `{"transaction_ids":["10"]}`, anonymous: true, status: 401},
		{name: "forbidden", body: `{"transaction_ids":["10"]}`, err: h.Forbidden, status: 403, calls: 1},
		{name: "commit failed", body: `{"transaction_ids":["10"]}`, err: errors.New("database secret"), status: 500, calls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &bulkStore{err: tc.err}
			req := httptest.NewRequest("POST", "/", strings.NewReader(tc.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()
			c := echo.New().NewContext(req, recorder)
			c.SetParamNames("householdId")
			c.SetParamValues("1")
			if !tc.anonymous {
				c.Set(router.ContextKeyUserNo, "2")
			}
			handler := New(store)
			var err error
			if tc.unshare {
				err = handler.BulkUnshare(c)
			} else {
				err = handler.BulkShare(c)
			}
			if err != nil {
				t.Fatal(err)
			}
			if recorder.Code != tc.status || store.calls != tc.calls {
				t.Fatalf("status=%d calls=%d body=%s", recorder.Code, store.calls, recorder.Body.String())
			}
			if tc.calls == 1 && store.share == tc.unshare {
				t.Fatal("wrong operation")
			}
			if tc.status == 200 {
				var response h.BulkSharingResponse
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				if len(response.Results) != 1 || response.Results[0].Outcome != "failed" {
					t.Fatal(response)
				}
			}
			if strings.Contains(recorder.Body.String(), "database secret") {
				t.Fatal("internal error exposed")
			}
		})
	}
}
