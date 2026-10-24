//go:build integration

package store_postgres

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	h "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type transactionListQueryCounter struct {
	logger.Interface
	queries int
}

func (c *transactionListQueryCounter) Trace(context.Context, time.Time, func() (string, int64), error) {
	c.queries++
}

func listTransactionsInOneQuery(t *testing.T, db *gorm.DB, month, sharing, cursor string) ([]model.V1Transaction, *string) {
	t.Helper()
	counter := &transactionListQueryCounter{Interface: logger.Default}
	store := NewTransactionStore(db.Session(&gorm.Session{Logger: counter}))
	rows, next, err := store.ListV1Transactions("2", month, sharing, cursor)
	if err != nil {
		t.Fatal(err)
	}
	if counter.queries != 1 {
		t.Fatalf("queries = %d, want 1", counter.queries)
	}
	if rows == nil {
		t.Fatal("transactions must be an empty slice rather than null")
	}
	return rows, next
}

func TestListV1TransactionsPaginationIntegration(t *testing.T) {
	db := legacyTestDB(t)
	if err := db.Exec(`INSERT INTO sub_category (sub_category_id, user_no, category_id, sub_category_name) VALUES (1, 2, 1, '外食')`).Error; err != nil {
		t.Fatal(err)
	}
	for i, count := range []int{0, 100, 500, 501, 1001} {
		t.Run(fmt.Sprintf("%d rows", count), func(t *testing.T) {
			month := fmt.Sprintf("2026-%02d-01", i+1)
			if err := db.Exec(`INSERT INTO "transaction" (user_no, transaction_name, transaction_amount, transaction_date, category_id, sub_category_id, fixed_flg)
				SELECT 2, 'ランチ', -1200, ?::date, 1, 1, false FROM generate_series(1, ?)`, month, count).Error; err != nil {
				t.Fatal(err)
			}
			var expectedIDs []string
			if err := db.Table("transaction").Select("CAST(transaction_id AS TEXT)").Where("transaction_date = ?", month).Order("transaction.transaction_id DESC").Scan(&expectedIDs).Error; err != nil {
				t.Fatal(err)
			}
			cursor, offset := "", 0
			for {
				rows, next := listTransactionsInOneQuery(t, db, month, "all", cursor)
				want := min(500, count-offset)
				if len(rows) != want {
					t.Fatalf("page at %d has %d rows, want %d", offset, len(rows), want)
				}
				for j, row := range rows {
					if row.TransactionId != expectedIDs[offset+j] {
						t.Fatalf("row %d = %s, want %s", offset+j, row.TransactionId, expectedIDs[offset+j])
					}
				}
				offset += len(rows)
				if offset == count {
					if next != nil {
						t.Fatalf("last page cursor = %q, want nil", *next)
					}
					break
				}
				if next == nil || *next != rows[len(rows)-1].TransactionId {
					t.Fatal("next cursor must be the last returned transaction ID")
				}
				cursor = *next
			}
		})
	}
}

func TestListV1TransactionsSharingPaginationIntegration(t *testing.T) {
	db := legacyTestDB(t)
	for _, statement := range []string{
		`INSERT INTO sub_category (sub_category_id, user_no, category_id, sub_category_name) VALUES (1, 2, 1, '外食')`,
		`INSERT INTO "transaction" (user_no, transaction_name, transaction_amount, transaction_date, category_id, sub_category_id, fixed_flg)
		 SELECT 2, 'ランチ', -1200, '2026-09-25', 1, 1, false FROM generate_series(1, 1001)`,
		`INSERT INTO household (household_id, name, state) VALUES (1, '家族', 'active')`,
		`INSERT INTO household_member (member_id, household_id, user_no, role, state, slot_no, display_name) VALUES (1, 1, 2, 'admin', 'active', 1, '本人')`,
		`INSERT INTO household_entry (household_id, kind, source_transaction_id, payer_member_id, state, created_by, updated_by)
		 SELECT 1, 'shared', transaction_id, 1, 'active', 1, 1 FROM "transaction" WHERE transaction_id % 2 = 0`,
		`INSERT INTO household_entry (household_id, kind, source_transaction_id, payer_member_id, state, created_by, updated_by)
		 VALUES (1, 'shared', 3, 1, 'withdrawn', 1, 1), (1, 'snapshot', 5, 1, 'active', 1, 1)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		sharing string
		count   int
		shared  bool
	}{{"shared", 500, true}, {"private", 501, false}} {
		t.Run(tc.sharing, func(t *testing.T) {
			cursor, seen := "", map[string]bool{}
			previous := uint64(1002)
			for {
				rows, next := listTransactionsInOneQuery(t, db, "2026-09-01", tc.sharing, cursor)
				if len(rows) == 0 {
					t.Fatal("unexpected empty page")
				}
				for _, row := range rows {
					id, err := strconv.ParseUint(row.TransactionId, 10, 64)
					if err != nil {
						t.Fatal(err)
					}
					if seen[row.TransactionId] || id >= previous || (id%2 == 0) != tc.shared || row.Shared != tc.shared {
						t.Fatalf("unexpected transaction: %+v", row)
					}
					seen[row.TransactionId], previous = true, id
				}
				if next == nil {
					break
				}
				cursor = *next
			}
			if len(seen) != tc.count {
				t.Fatalf("total rows = %d, want %d", len(seen), tc.count)
			}
		})
	}
}

func TestListV1TransactionsResourcesAndValidationIntegration(t *testing.T) {
	db := legacyTestDB(t)
	for _, statement := range []string{
		`INSERT INTO sub_category (sub_category_id, user_no, category_id, sub_category_name) VALUES (1, 2, 1, '外食')`,
		`INSERT INTO payment_resource (payment_id, user_no, payment_name) VALUES (11, 2, '現金')`,
		`INSERT INTO "transaction" (transaction_id, user_no, transaction_name, transaction_amount, transaction_date, transaction_time, category_id, sub_category_id, fixed_flg, payment_id, version, deleted_at) VALUES
		 (1, 2, '支出', -1200, '2026-09-01', '12:34:56', 1, 1, true, 11, 3, NULL),
		 (2, 2, '収入', 2500, '2026-09-30', NULL, 1, 1, false, NULL, 1, NULL),
		 (3, 2, 'ゼロ', 0, '2026-09-15', NULL, 1, 1, false, NULL, 1, NULL),
		 (4, 2, '削除済み', -100, '2026-09-15', NULL, 1, 1, false, NULL, 1, CURRENT_TIMESTAMP),
		 (5, 3, '他の人', -100, '2026-09-15', NULL, 1, 1, false, NULL, 1, NULL),
		 (6, 2, '前月', -100, '2026-08-31', NULL, 1, 1, false, NULL, 1, NULL),
		 (7, 2, '翌月', -100, '2026-10-01', NULL, 1, 1, false, NULL, 1, NULL)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	rows, next := listTransactionsInOneQuery(t, db, "2026-09-01", "", "")
	if len(rows) != 3 || next != nil {
		t.Fatalf("rows = %d, next = %v", len(rows), next)
	}
	clock, paymentID, paymentName := "12:34", "11", "現金"
	expected := []model.V1Transaction{
		{Version: 1, TransactionId: "3", TransactionDate: "2026-09-15", TransactionName: "ゼロ", Sign: -1, CategoryId: "1", CategoryName: "食費", SubCategoryId: "1", SubCategoryName: "外食"},
		{Version: 1, TransactionId: "2", TransactionDate: "2026-09-30", TransactionName: "収入", Amount: 2500, Sign: 1, SignedAmount: 2500, CategoryId: "1", CategoryName: "食費", SubCategoryId: "1", SubCategoryName: "外食"},
		{Version: 3, TransactionId: "1", TransactionDate: "2026-09-01", TransactionTime: &clock, TransactionName: "支出", Amount: 1200, Sign: -1, SignedAmount: -1200, CategoryId: "1", CategoryName: "食費", SubCategoryId: "1", SubCategoryName: "外食", FixedFlg: true, PaymentId: &paymentID, PaymentName: &paymentName},
	}
	if !reflect.DeepEqual(rows, expected) {
		t.Fatalf("rows = %+v, want %+v", rows, expected)
	}
	for _, row := range rows {
		detail, err := NewTransactionStore(db).GetV1Transaction("2", row.TransactionId)
		if err != nil || !reflect.DeepEqual(*detail, row) {
			t.Fatalf("list and detail differ for %s: %v", row.TransactionId, err)
		}
	}
	for _, tc := range []struct{ month, sharing, cursor string }{
		{"invalid", "all", ""}, {"2026-09-02", "all", ""}, {"2026-09-01", "invalid", ""}, {"2026-09-01", "all", "invalid"},
	} {
		_, _, err := NewTransactionStore(db).ListV1Transactions("2", tc.month, tc.sharing, tc.cursor)
		if !errors.Is(err, h.Invalid) {
			t.Fatalf("query %+v: error = %v, want invalid", tc, err)
		}
	}
}
