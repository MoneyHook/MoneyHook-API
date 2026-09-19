package store_postgres

import (
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestV1TransactionTimeSelectUsesCockroachCompatibleExpression(t *testing.T) {
	want := "LEFT(CAST(t.transaction_time AS TEXT), 5) AS transaction_time"
	if v1TransactionTimeSelect != want {
		t.Fatalf("v1TransactionTimeSelect = %q, want %q", v1TransactionTimeSelect, want)
	}
}

func TestFrequentTransactionNameQueryDeduplicatesAndLimits(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "user=test"}), &gorm.Config{
		DisableAutomaticPing: true,
		DryRun:               true,
	})
	if err != nil {
		t.Fatalf("open dry-run database: %v", err)
	}

	query := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return frequentTransactionNameQuery(tx, "user-1", 20, "2026-09-19").Find(&[]frequentTransactionNameRecord{})
	})

	for _, fragment := range []string{
		"ROW_NUMBER() OVER (PARTITION BY transaction_name",
		"row_num = 1",
		"ORDER BY name_score DESC, name_last_used_date DESC, name_usage_count DESC, transaction_name ASC",
		"LIMIT 20",
	} {
		if !strings.Contains(query, fragment) {
			t.Errorf("query does not contain %q:\n%s", fragment, query)
		}
	}
}

type frequentTransactionNameRecord struct{}

func (frequentTransactionNameRecord) TableName() string { return "frequent_transactions" }

func TestFrequentTransactionReferenceDate(t *testing.T) {
	for _, tc := range []struct{ instant, want string }{
		{"2026-09-18T14:59:59Z", "2026-09-18"},
		{"2026-09-18T15:00:00Z", "2026-09-19"},
	} {
		instant, err := time.Parse(time.RFC3339, tc.instant)
		if err != nil {
			t.Fatal(err)
		}
		if got := frequentTransactionReferenceDate(instant); got != tc.want {
			t.Fatalf("got %s, want %s", got, tc.want)
		}
	}
}
