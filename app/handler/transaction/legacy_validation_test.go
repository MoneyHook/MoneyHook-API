package transaction

import (
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	"MoneyHook/MoneyHook-API/router"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type legacyWriteStore struct {
	transactiondomain.Store
	calls int
	input *model.AddTransactionList
}

func (s *legacyWriteStore) AddTransactionList(input *model.AddTransactionList) error {
	s.calls++
	s.input = input
	return nil
}
func (s *legacyWriteStore) AddTransaction(input *model.AddTransaction) error   { s.calls++; return nil }
func (s *legacyWriteStore) EditTransaction(input *model.EditTransaction) error { s.calls++; return nil }

func validLegacyInput() map[string]any {
	return map[string]any{"transaction_date": "2026-09-01", "transaction_amount": 1200, "transaction_sign": -1, "transaction_name": "ランチ", "category_id": "2", "sub_category_id": "4", "fixed_flg": false, "transaction_id": "7"}
}

func TestLegacyWritesRejectInvalidInputBeforePersistence(t *testing.T) {
	message.Read()
	cases := []struct {
		name   string
		change func(map[string]any)
	}{
		{"missing amount", func(v map[string]any) { delete(v, "transaction_amount") }},
		{"null amount", func(v map[string]any) { v["transaction_amount"] = nil }},
		{"negative amount", func(v map[string]any) { v["transaction_amount"] = -1 }},
		{"fractional amount", func(v map[string]any) { v["transaction_amount"] = 1.5 }},
		{"amount overflow", func(v map[string]any) { v["transaction_amount"] = json.Number("9223372036854775808") }},
		{"invalid sign", func(v map[string]any) { v["transaction_sign"] = 2 }},
		{"missing sign", func(v map[string]any) { delete(v, "transaction_sign") }},
		{"invalid date", func(v map[string]any) { v["transaction_date"] = "2026-02-30" }},
		{"empty name", func(v map[string]any) { v["transaction_name"] = " " }},
		{"long name", func(v map[string]any) { v["transaction_name"] = strings.Repeat("あ", 33) }},
		{"category ID", func(v map[string]any) { v["category_id"] = "no" }},
		{"subcategory ID", func(v map[string]any) { v["sub_category_id"] = "0" }},
		{"subcategory name required", func(v map[string]any) { delete(v, "sub_category_id") }},
		{"long subcategory name", func(v map[string]any) { v["sub_category_id"] = ""; v["sub_category_name"] = strings.Repeat("あ", 17) }},
		{"payment ID", func(v map[string]any) { v["payment_id"] = "-1" }},
		{"missing false flag", func(v map[string]any) { delete(v, "fixed_flg") }},
		{"null flag", func(v map[string]any) { v["fixed_flg"] = nil }},
	}
	for _, tc := range cases {
		for _, operation := range []string{"add", "batch", "edit"} {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				input := validLegacyInput()
				tc.change(input)
				body := map[string]any{"transaction": input}
				if operation == "batch" {
					body = map[string]any{"transaction_list": []any{validLegacyInput(), input}}
				}
				encoded, _ := json.Marshal(body)
				store := &legacyWriteStore{}
				h := New(store, nil)
				handler := h.AddTransaction
				if operation == "batch" {
					handler = h.AddTransactionList
				}
				if operation == "edit" {
					handler = h.EditTransaction
				}
				rec := runLegacyHandler(t, handler, string(encoded))
				if rec.Code != 422 || store.calls != 0 {
					t.Fatalf("status=%d store calls=%d body=%s", rec.Code, store.calls, rec.Body.String())
				}
			})
		}
	}
}

func runLegacyHandler(t *testing.T, handler echo.HandlerFunc, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/?month=2026-09-01", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.Set(router.ContextKeyUserNo, "2")
	c.SetParamNames("transactionId")
	c.SetParamValues("7")
	if err := handler(c); err != nil {
		t.Fatal(err)
	}
	return rec
}

func TestLegacyBatchValidationPreservesZeroAndFalse(t *testing.T) {
	store := &legacyWriteStore{}
	h := New(store, nil)
	input := validLegacyInput()
	input["transaction_amount"] = 0
	input["sub_category_id"] = ""
	input["sub_category_name"] = "新カテゴリ"
	encoded, _ := json.Marshal(map[string]any{"transaction_list": []any{input, validLegacyInput()}})
	rec := runLegacyHandler(t, h.AddTransactionList, string(encoded))
	if rec.Code != 200 || store.calls != 1 {
		t.Fatalf("status=%d calls=%d", rec.Code, store.calls)
	}
	if store.input.UserId != "2" || store.input.TransactionList[0].FixedFlg || store.input.TransactionList[0].TransactionAmount != 0 || store.input.TransactionList[1].TransactionAmount != -1200 {
		t.Fatalf("unexpected input: %+v", store.input)
	}
}

func TestLegacyBatchRejectsEmptyAndMalformedRequests(t *testing.T) {
	message.Read()
	for _, body := range []string{`{}`, `{"transaction_list":[]}`, `{"transaction_list":null}`, `{`} {
		store := &legacyWriteStore{}
		rec := runLegacyHandler(t, New(store, nil).AddTransactionList, body)
		if rec.Code != 422 || store.calls != 0 {
			t.Fatalf("body=%s status=%d calls=%d", body, rec.Code, store.calls)
		}
	}
}

func TestLegacyEditRequiresTransactionID(t *testing.T) {
	message.Read()
	for _, id := range []string{"", "0", "bad"} {
		input := validLegacyInput()
		input["transaction_id"] = id
		encoded, _ := json.Marshal(map[string]any{"transaction": input})
		store := &legacyWriteStore{}
		rec := runLegacyHandler(t, New(store, nil).EditTransaction, string(encoded))
		if rec.Code != 422 || store.calls != 0 {
			t.Fatalf("id=%s status=%d calls=%d", id, rec.Code, store.calls)
		}
	}
}
