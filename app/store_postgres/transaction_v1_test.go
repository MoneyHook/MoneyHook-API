package store_postgres

import "testing"

func TestV1TransactionTimeSelectUsesCockroachCompatibleExpression(t *testing.T) {
	want := "LEFT(CAST(t.transaction_time AS TEXT), 5) AS transaction_time"
	if v1TransactionTimeSelect != want {
		t.Fatalf("v1TransactionTimeSelect = %q, want %q", v1TransactionTimeSelect, want)
	}
}
