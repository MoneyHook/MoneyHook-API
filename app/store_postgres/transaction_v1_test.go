package store_postgres

import (
	"strings"
	"testing"

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
		return frequentTransactionNameQuery(tx, "user-1", 20).Find(&[]frequentTransactionNameRecord{})
	})

	for _, fragment := range []string{
		"ROW_NUMBER() OVER (PARTITION BY tran.transaction_name",
		"row_num = 1",
		"ORDER BY usage_count DESC, transaction_name ASC",
		"LIMIT 20",
	} {
		if !strings.Contains(query, fragment) {
			t.Errorf("query does not contain %q:\n%s", fragment, query)
		}
	}
}

type frequentTransactionNameRecord struct{}

func (frequentTransactionNameRecord) TableName() string { return "frequent_transactions" }
