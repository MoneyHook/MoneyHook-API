//go:build integration

package store_postgres

import (
	"fmt"
	"math"
	"testing"
	"time"

	"MoneyHook/MoneyHook-API/model"
	"gorm.io/gorm"
)

func frequentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := legacyTestDB(t)
	for _, sql := range []string{
		`CREATE TABLE category (category_id BIGINT PRIMARY KEY, category_name TEXT NOT NULL)`,
		`INSERT INTO category VALUES (1, 'Food'), (2, 'Other')`,
		`INSERT INTO sub_category (sub_category_id, user_no, category_id, sub_category_name) VALUES (1, 2, 1, 'Lunch'), (2, 2, 2, 'Other')`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func addFrequentRows(t *testing.T, db *gorm.DB, name string, days, count, category int, payment any, fixed bool, user int) {
	t.Helper()
	err := db.Exec(`INSERT INTO "transaction" (user_no, transaction_name, transaction_amount, transaction_date, category_id, sub_category_id, fixed_flg, payment_id)
 SELECT ?, ?, -100, '2026-09-19'::date - ?::integer, ?, ?, ?, ? FROM generate_series(1, ?)`, user, name, days, category, category, fixed, payment, count).Error
	if err != nil {
		t.Fatal(err)
	}
}

func frequentResults(t *testing.T, db *gorm.DB, limit int) []model.FrequentTransactionName {
	t.Helper()
	var rows []model.FrequentTransactionName
	if err := frequentTransactionNameQuery(db, "2", limit, "2026-09-19").Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	return rows
}

func TestFrequentTransactionRankingIntegration(t *testing.T) {
	t.Run("weights", func(t *testing.T) {
		db := frequentTestDB(t)
		for i, days := range []int{0, 60, 120} {
			addFrequentRows(t, db, fmt.Sprint(i), days, 1, 1, nil, false, 2)
		}
		var rows []struct{ NameScore float64 }
		if err := frequentTransactionNameQuery(db, "2", 20, "2026-09-19").Scan(&rows).Error; err != nil {
			t.Fatal(err)
		}
		if len(rows) != 3 {
			t.Fatalf("rows: %+v", rows)
		}
		for i, want := range []float64{1, 0.5, 0.25} {
			if math.Abs(rows[i].NameScore-want) > 1e-12 {
				t.Fatalf("score: %v, want %v", rows[i], want)
			}
		}
	})
	t.Run("aggregate names and select real recent configuration", func(t *testing.T) {
		db := frequentTestDB(t)
		addFrequentRows(t, db, "split", 0, 2, 2, 22, true, 2)
		addFrequentRows(t, db, "split", 120, 5, 1, 11, false, 2)
		addFrequentRows(t, db, "competitor", 0, 3, 1, nil, false, 2)
		addFrequentRows(t, db, "old", 600, 100, 1, nil, false, 2)
		addFrequentRows(t, db, "future", -1, 100, 1, nil, false, 2)
		addFrequentRows(t, db, "other user", 0, 100, 1, nil, false, 3)
		rows := frequentResults(t, db, 20)
		if len(rows) != 3 || rows[0].TransactionName != "split" || rows[1].TransactionName != "competitor" || rows[2].TransactionName != "old" {
			t.Fatalf("ranking: %+v", rows)
		}
		if rows[0].CategoryId != "2" || rows[0].SubCategoryId != "2" || rows[0].PaymentId != "22" || !rows[0].FixedFlg {
			t.Fatalf("configuration: %+v", rows[0])
		}
		if rows[1].PaymentId != "" {
			t.Fatalf("null payment: %+v", rows[1])
		}
	})
	t.Run("name ties use date then count then name", func(t *testing.T) {
		db := frequentTestDB(t)
		addFrequentRows(t, db, "old", 60, 2, 1, nil, false, 2)
		addFrequentRows(t, db, "recent", 0, 1, 1, nil, false, 2)
		addFrequentRows(t, db, "a", 0, 1, 1, nil, false, 2)
		addFrequentRows(t, db, "a", 60, 2, 1, nil, false, 2)
		addFrequentRows(t, db, "b", 0, 2, 1, nil, false, 2)
		addFrequentRows(t, db, "c", 0, 2, 1, nil, false, 2)
		rows := frequentResults(t, db, 20)
		for i, want := range []string{"a", "b", "c", "recent", "old"} {
			if len(rows) != 5 || rows[i].TransactionName != want {
				t.Fatalf("ties: %+v", rows)
			}
		}
	})
	t.Run("configuration ties", func(t *testing.T) {
		db := frequentTestDB(t)
		addFrequentRows(t, db, "date", 60, 2, 1, 11, false, 2)
		addFrequentRows(t, db, "date", 0, 1, 2, 22, true, 2)
		addFrequentRows(t, db, "count", 0, 2, 1, 11, false, 2)
		addFrequentRows(t, db, "count", 0, 1, 2, 22, true, 2)
		addFrequentRows(t, db, "count", 60, 2, 2, 22, true, 2)
		for _, p := range []any{nil, 11, 22} {
			for _, fixed := range []bool{false, true} {
				for _, cat := range []int{1, 2} {
					addFrequentRows(t, db, "ids", 0, 1, cat, p, fixed, 2)
				}
			}
		}
		addFrequentRows(t, db, "subcategory", 0, 1, 1, 11, false, 2)
		addFrequentRows(t, db, "subcategory", 0, 1, 1, 22, false, 2)
		if err := db.Exec(`INSERT INTO sub_category (sub_category_id, user_no, category_id, sub_category_name) VALUES (3, 2, 1, 'Dinner')`).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec(`UPDATE "transaction" SET sub_category_id=3 WHERE transaction_name='subcategory' AND payment_id=11`).Error; err != nil {
			t.Fatal(err)
		}
		addFrequentRows(t, db, "fixed", 0, 1, 1, 11, true, 2)
		addFrequentRows(t, db, "fixed", 0, 1, 1, 22, false, 2)
		addFrequentRows(t, db, "payment", 0, 1, 1, 22, false, 2)
		addFrequentRows(t, db, "payment", 0, 1, 1, 11, false, 2)
		for _, r := range frequentResults(t, db, 20) {
			switch r.TransactionName {
			case "date", "count":
				if r.CategoryId != "2" || r.PaymentId != "22" {
					t.Fatalf("tie: %+v", r)
				}
			case "subcategory", "fixed":
				if r.SubCategoryId != "1" || r.FixedFlg || r.PaymentId != "22" {
					t.Fatalf("tie: %+v", r)
				}
			case "payment":
				if r.PaymentId != "11" {
					t.Fatalf("tie: %+v", r)
				}
			case "ids":
				if r.CategoryId != "1" || r.SubCategoryId != "1" || r.FixedFlg || r.PaymentId != "" {
					t.Fatalf("IDs: %+v", r)
				}
			}
		}
	})
	t.Run("empty and limits after deduplication", func(t *testing.T) {
		db := frequentTestDB(t)
		if rows := frequentResults(t, db, 20); len(rows) != 0 {
			t.Fatal(rows)
		}
		for i := 0; i < 105; i++ {
			for _, p := range []any{nil, 11} {
				addFrequentRows(t, db, fmt.Sprintf("name-%03d", i), 0, 1, 1, p, false, 2)
			}
		}
		for _, limit := range []int{20, 100} {
			rows := frequentResults(t, db, limit)
			if len(rows) != limit {
				t.Fatalf("got %d want %d", len(rows), limit)
			}
			for i, r := range rows {
				if r.TransactionName != fmt.Sprintf("name-%03d", i) {
					t.Fatal(rows)
				}
			}
		}
	})
}

func TestFrequentTransactionQueryPerformance(t *testing.T) {
	db := frequentTestDB(t)
	if err := db.Exec(`INSERT INTO "transaction" (user_no, transaction_name, transaction_amount, transaction_date, category_id, sub_category_id, fixed_flg, payment_id)
 SELECT 2, 'name-' || (i % 200)::text, -100, '2026-09-19'::date - (i % 730), 1, 1, false, i % 3 FROM generate_series(1, 20000) AS g(i)`).Error; err != nil {
		t.Fatal(err)
	}
	old := `SELECT * FROM (SELECT tran.transaction_name, tran.category_id, c.category_name, tran.sub_category_id, sc.sub_category_name, tran.fixed_flg, tran.payment_id, COUNT(*) AS usage_count,
 ROW_NUMBER() OVER (PARTITION BY tran.transaction_name ORDER BY COUNT(*) DESC, tran.category_id, tran.sub_category_id, tran.fixed_flg, tran.payment_id NULLS FIRST) AS row_num
 FROM "transaction" tran JOIN category c ON c.category_id=tran.category_id JOIN sub_category sc ON sc.sub_category_id=tran.sub_category_id WHERE tran.user_no=2
 GROUP BY tran.transaction_name, tran.category_id, c.category_name, tran.sub_category_id, sc.sub_category_name, tran.fixed_flg, tran.payment_id
 ORDER BY usage_count DESC, tran.transaction_name, tran.category_id, tran.sub_category_id, tran.fixed_flg, tran.payment_id NULLS FIRST) AS frequent_transactions
 WHERE row_num=1 ORDER BY usage_count DESC, transaction_name LIMIT 20`
	next := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return frequentTransactionNameQuery(tx, "2", 20, "2026-09-19").Find(&[]model.FrequentTransactionName{})
	})
	for _, tc := range []struct{ name, sql string }{{"old", old}, {"new", next}} {
		start := time.Now()
		for i := 0; i < 5; i++ {
			var rows []model.FrequentTransactionName
			if err := db.Raw(tc.sql).Scan(&rows).Error; err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("%s 20,000 transactions / 200 names: mean %s", tc.name, time.Since(start)/5)
		var plan []string
		if err := db.Raw("EXPLAIN ANALYZE " + tc.sql).Scan(&plan).Error; err != nil {
			t.Fatal(err)
		}
		for _, line := range plan {
			t.Log(line)
		}
	}
}
