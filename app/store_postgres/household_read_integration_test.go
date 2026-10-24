//go:build integration

package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"reflect"
	"testing"
	"time"
)

type householdReadCounter struct {
	logger.Interface
	count int
}

func (c *householdReadCounter) Trace(context.Context, time.Time, func() (string, int64), error) {
	c.count++
}

func householdReadFixture(t *testing.T) (*gorm.DB, *HouseholdStore, *HouseholdStore, *householdReadCounter, uint64, uint64, uint64) {
	t.Helper()
	db := legacyTestDB(t)
	for _, stmt := range []string{
		`INSERT INTO sub_category (sub_category_id,user_no,category_id,sub_category_name) VALUES (1,2,1,'外食')`,
		`INSERT INTO household (household_id,name,state) VALUES (1,'家族','active'),(2,'別の家族','active')`,
		`INSERT INTO household_member (member_id,household_id,user_no,role,state,slot_no,display_name) VALUES (1,1,2,'admin','active',1,'本人'),(2,1,3,'member','active',2,'現在の名前')`,
		`INSERT INTO household_payment (payment_id,household_id,name,payment_type_id) VALUES (1,1,'家族カード',1)`,
		`INSERT INTO household_sub_category (sub_category_id,household_id,category_id,name) VALUES (1,1,1,'家族外食')`,
		`INSERT INTO "transaction" (user_no,transaction_name,transaction_amount,transaction_date,transaction_time,category_id,sub_category_id,fixed_flg)
 SELECT 2,'昼食',-100,'2026-09-10','12:34:56',1,1,false FROM generate_series(1,1001)`,
		`INSERT INTO "transaction" (user_no,transaction_name,transaction_amount,transaction_date,category_id,sub_category_id,fixed_flg) VALUES
 (2,'収入',1000,'2026-09-10',1,1,false),(2,'ゼロ',0,'2026-09-10',1,1,false)`,
		`INSERT INTO household_entry (household_id,kind,source_transaction_id,payer_member_id,payment_id,sub_category_id,created_by,updated_by)
 SELECT 1,'shared',transaction_id,1,1,1,1,1 FROM "transaction"`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	store := NewHouseholdStore(db, "read-query-test-secret-32-characters")
	// Stored snapshots retain their captured payer; corrections can move them between months.
	payerID := "2"
	addPayload := func(kind string, payload, corrected h.Entry, excluded bool) uint64 {
		t.Helper()
		var id uint64
		state := "active"
		if excluded {
			state = "withdrawn"
		}
		source := any(nil)
		if kind == "snapshot" {
			source = 1
		}
		// A different source ID is required for each snapshot because of the family/source unique constraint.
		if kind == "snapshot" {
			var next uint64
			if err := db.Raw(`INSERT INTO "transaction" (user_no,transaction_name,transaction_amount,transaction_date,category_id,sub_category_id,fixed_flg) VALUES (2,'控えの原本',-1,'2026-08-01',1,1,false) RETURNING transaction_id`).Scan(&next).Error; err != nil {
				t.Fatal(err)
			}
			source = next
		}
		if err := db.Raw(`INSERT INTO household_entry (household_id,kind,source_transaction_id,state,created_by,updated_by) VALUES (1,?,?,?,1,1) RETURNING entry_id`, kind, source, state).Scan(&id).Error; err != nil {
			t.Fatal(err)
		}
		var correction any
		if corrected.Name != "" {
			correction = string(publicEntry(&corrected))
		}
		if err := db.Exec(`INSERT INTO household_entry_data (entry_id,payload,corrected_payload,captured_at) VALUES (?,?::jsonb,?::jsonb,CURRENT_TIMESTAMP)`, id, string(publicEntry(&payload)), correction).Error; err != nil {
			t.Fatal(err)
		}
		return id
	}
	payload := func(name, date string, amount int64, sign int) h.Entry {
		return h.Entry{TransactionInput: h.TransactionInput{Name: name, Date: date, Amount: amount, Sign: sign, CategoryID: "2"}, CategoryName: "日用品", SignedAmount: amount * int64(sign), Payer: h.Payer{Kind: "member", MemberID: &payerID, DisplayName: "保存時の名前", State: "left"}}
	}
	proxyID := addPayload("proxy", payload("代理支出", "2026-09-11", 200, -1), h.Entry{}, false)
	addPayload("proxy", payload("代理収入", "2026-09-11", 300, 1), h.Entry{}, false)
	snapshotID := addPayload("snapshot", payload("元の控え", "2026-08-01", 50, -1), payload("訂正済み", "2026-09-12", 400, -1), false)
	excluded := payload("除外", "2026-09-12", 9999, -1)
	excluded.Excluded = true
	addPayload("snapshot", excluded, h.Entry{}, false)
	movedID := addPayload("snapshot", payload("移動前", "2026-09-12", 50, -1), payload("翌月へ移動", "2026-10-01", 850, -1), false)
	addPayload("proxy", payload("取り下げ", "2026-09-11", 9000, -1), h.Entry{}, true)
	// A deleted shared original must not reappear, while snapshots survive original deletion.
	if err := db.Exec(`WITH original AS (
 INSERT INTO "transaction" (user_no,transaction_name,transaction_amount,transaction_date,category_id,sub_category_id,fixed_flg,deleted_at)
 VALUES (2,'削除済み',-100,'2026-09-10',1,1,false,CURRENT_TIMESTAMP) RETURNING transaction_id)
 INSERT INTO household_entry (household_id,kind,source_transaction_id,payer_member_id,created_by,updated_by)
 SELECT 1,'shared',transaction_id,1,1,1 FROM original`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`UPDATE "transaction" SET deleted_at=CURRENT_TIMESTAMP WHERE transaction_id IN (SELECT source_transaction_id FROM household_entry WHERE kind='snapshot')`).Error; err != nil {
		t.Fatal(err)
	}
	counter := &householdReadCounter{Interface: logger.Default}
	counted := NewHouseholdStore(db.Session(&gorm.Session{Logger: counter}), "read-query-test-secret-32-characters")
	return db, store, counted, counter, proxyID, snapshotID, movedID
}

func TestHouseholdReadQueriesIntegration(t *testing.T) {
	_, store, counted, counter, proxyID, snapshotID, movedID := householdReadFixture(t)
	ctx := context.Background()
	t.Run("pagination and SQL count", func(t *testing.T) {
		cursor := ""
		seen := map[string]bool{}
		pages := 0
		for {
			counter.count = 0
			p, err := counted.Entries(ctx, "2", "1", h.Filter{Month: "2026-09-01", Cursor: cursor})
			if err != nil {
				t.Fatal(err)
			}
			if counter.count != 3 {
				t.Fatalf("queries = %d, want 3", counter.count)
			}
			if len(p.Entries) > 100 {
				t.Fatal("page limit changed")
			}
			pages++
			for _, entry := range p.Entries {
				if seen[entry.ID] {
					t.Fatal("duplicate entry")
				}
				seen[entry.ID] = true
				if entry.PaymentID != nil || entry.SubCategoryID != "" || entry.SubCategoryName != "" {
					t.Fatal("personal references exposed")
				}
				if entry.ID == "1" || entry.ID == stringID(proxyID) || entry.ID == stringID(snapshotID) {
					detail, e := store.Entry(ctx, "2", "1", entry.ID)
					if e != nil || !reflect.DeepEqual(*detail, entry) {
						t.Fatalf("list/detail mismatch: %+v %v", entry, e)
					}
				}
			}
			if p.NextCursor == nil {
				break
			}
			cursor = *p.NextCursor
		}
		if len(seen) != 1007 || pages != 11 || seen[stringID(movedID)] {
			t.Fatalf("rows=%d pages=%d", len(seen), pages)
		}
	})
	t.Run("projection privacy and captured names", func(t *testing.T) {
		row, err := store.Entry(ctx, "3", "1", "1")
		if err != nil {
			t.Fatal(err)
		}
		if row.SourceTransactionID != nil || row.SourceVersion != nil || row.Permissions.Edit || row.HouseholdPaymentName == nil || *row.HouseholdPaymentName != "家族カード" || row.Time == nil || *row.Time != "12:34" {
			t.Fatalf("unexpected projection: %+v", row)
		}
		snap, err := store.Entry(ctx, "2", "1", stringID(snapshotID))
		if err != nil {
			t.Fatal(err)
		}
		if !snap.Corrected || snap.Payer.DisplayName != "保存時の名前" || snap.CapturedAt == nil || !snap.Permissions.Correct || snap.Amount != 400 {
			t.Fatalf("bad snapshot: %+v", snap)
		}
	})
}
