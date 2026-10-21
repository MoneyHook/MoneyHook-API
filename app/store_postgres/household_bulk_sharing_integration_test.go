//go:build integration

package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestHouseholdBulkSharingLifecycle(t *testing.T) {
	db, s, family := householdReferenceFixture(t)
	ctx := context.Background()
	ts := NewTransactionStore(db)
	input := v1Row()
	input.PaymentId = ptr("10")
	first, err := ts.CreateV1Transaction(&input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := ts.CreateV1Transaction(&input)
	if err != nil {
		t.Fatal(err)
	}
	input.UserId = "3"
	foreign, err := ts.CreateV1Transaction(&input)
	// User 3 cannot use user 2's payment.
	if err == nil {
		t.Fatal("foreign payment accepted", foreign)
	}
	input.PaymentId = nil
	foreign, err = ts.CreateV1Transaction(&input)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := s.BulkShare(ctx, "2", family, h.BulkSharingInput{TransactionIDs: []string{second.TransactionId, foreign.TransactionId}}, false)
	if err != nil {
		t.Fatal(err)
	}
	assertBulkOutcomes(t, initial, "already-unshared", "failed")
	request := h.BulkSharingInput{TransactionIDs: []string{first.TransactionId, foreign.TransactionId, second.TransactionId, first.TransactionId, "999999"}}
	out, err := s.BulkShare(ctx, "2", family, request, true)
	if err != nil {
		t.Fatal(err)
	}
	assertBulkOutcomes(t, out, "shared", "failed", "shared", "failed")
	for _, i := range []int{1, 3} {
		if out.Results[i].Code != "NOT_FOUND" {
			t.Fatal(out)
		}
	}
	entries, err := s.Entries(ctx, "2", family, h.Filter{Month: "2026-09-01"})
	if err != nil || len(entries.Entries) != 2 {
		t.Fatalf("entries=%+v err=%v", entries, err)
	}
	for _, entry := range entries.Entries {
		if entry.Version != 1 || entry.HouseholdSubCategoryName == nil || *entry.HouseholdSubCategoryName != "外食" || entry.HouseholdPaymentName == nil || *entry.HouseholdPaymentName != "カード" {
			t.Fatal(entry)
		}
	}
	retry, err := s.BulkShare(ctx, "2", family, request, true)
	if err != nil {
		t.Fatal(err)
	}
	assertBulkOutcomes(t, retry, "already-shared", "failed", "already-shared", "failed")
	status, err := s.ShareStatus(ctx, "2", family, first.TransactionId)
	if err != nil || status.Version != 1 {
		t.Fatalf("retry changed version: %+v %v", status, err)
	}
	request.TransactionIDs = []string{first.TransactionId, second.TransactionId}
	out, err = s.BulkShare(ctx, "2", family, request, false)
	if err != nil {
		t.Fatal(err)
	}
	assertBulkOutcomes(t, out, "unshared", "unshared")
	out, err = s.BulkShare(ctx, "2", family, request, false)
	if err != nil {
		t.Fatal(err)
	}
	assertBulkOutcomes(t, out, "already-unshared", "already-unshared")
	out, err = s.BulkShare(ctx, "2", family, request, true)
	if err != nil {
		t.Fatal(err)
	}
	assertBulkOutcomes(t, out, "shared", "shared")
	status, err = s.ShareStatus(ctx, "2", family, first.TransactionId)
	if err != nil || status.Version != 3 {
		t.Fatalf("reshare: %+v %v", status, err)
	}
	if _, err := ts.GetV1Transaction("2", first.TransactionId); err != nil {
		t.Fatal("original removed", err)
	}
	if err := db.Model(&householdEntryRecord{}).Where("source_transaction_id = ?", first.TransactionId).Update("kind", "snapshot").Error; err != nil {
		t.Fatal(err)
	}
	for _, share := range []bool{true, false} {
		out, err = s.BulkShare(ctx, "2", family, h.BulkSharingInput{TransactionIDs: []string{first.TransactionId}}, share)
		if err != nil {
			t.Fatal(err)
		}
		assertBulkOutcomes(t, out, "failed")
		if out.Results[0].Code != "SOURCE_ALREADY_SNAPSHOTTED" {
			t.Fatal(out)
		}
	}
	if out, err = s.BulkShare(ctx, "3", family, request, true); !errors.Is(err, h.NotFound) || out != nil {
		t.Fatalf("nonmember: %+v %v", out, err)
	}
	if err := db.Model(&h.Family{}).Where("household_id = ?", family).Update("state", "archived").Error; err != nil {
		t.Fatal(err)
	}
	if out, err = s.BulkShare(ctx, "2", family, request, true); !errors.Is(err, h.Forbidden) || out != nil {
		t.Fatalf("archived: %+v %v", out, err)
	}
}

func TestHouseholdBulkSharingRollsBackFailedItemReferences(t *testing.T) {
	db, s, family := householdReferenceFixture(t)
	ts := NewTransactionStore(db)
	input := v1Row()
	good, err := ts.CreateV1Transaction(&input)
	if err != nil {
		t.Fatal(err)
	}
	input.SubCategoryName = "失敗用分類"
	input.PaymentId = ptr("10")
	bad, err := ts.CreateV1Transaction(&input)
	if err != nil {
		t.Fatal(err)
	}
	input.SubCategoryName = "成功用分類"
	input.PaymentId = nil
	last, err := ts.CreateV1Transaction(&input)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(fmt.Sprintf("ALTER TABLE household_entry ADD CONSTRAINT reject_item CHECK (source_transaction_id <> %s)", bad.TransactionId)).Error; err != nil {
		t.Fatal(err)
	}
	request := h.BulkSharingInput{TransactionIDs: []string{good.TransactionId, bad.TransactionId, last.TransactionId}}
	out, err := s.BulkShare(context.Background(), "2", family, request, true)
	if err != nil {
		t.Fatal(err)
	}
	assertBulkOutcomes(t, out, "shared", "failed", "shared")
	if out.Results[1].Code != "INTERNAL_ERROR" {
		t.Fatal(out)
	}
	var count int64
	if err := db.Table("household_sub_category").Where("name = ?", "失敗用分類").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("leaked subcategory: %d %v", count, err)
	}
	assertLegacyCount(t, db, "household_payment", 0)
	assertLegacyCount(t, db, "household_entry", 2)
	if out, err := s.BulkShare(context.Background(), "2", family, h.BulkSharingInput{TransactionIDs: []string{good.TransactionId, "invalid"}}, false); !errors.Is(err, h.Invalid) || out != nil {
		t.Fatalf("invalid batch: %+v %v", out, err)
	}
	status, err := s.ShareStatus(context.Background(), "2", family, good.TransactionId)
	if err != nil || status.State != "active" {
		t.Fatalf("invalid batch wrote changes: %+v %v", status, err)
	}
}

func TestHouseholdBulkSharingCommitFailure(t *testing.T) {
	db, s, family := householdReferenceFixture(t)
	input := v1Row()
	original, err := NewTransactionStore(db).CreateV1Transaction(&input)
	if err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		`CREATE FUNCTION fail_bulk_commit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'commit failure'; END $$`,
		`CREATE CONSTRAINT TRIGGER fail_commit AFTER INSERT ON household_entry DEFERRABLE INITIALLY DEFERRED FOR EACH ROW EXECUTE FUNCTION fail_bulk_commit()`,
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	out, err := s.BulkShare(context.Background(), "2", family, h.BulkSharingInput{TransactionIDs: []string{original.TransactionId}}, true)
	if err == nil || out != nil {
		t.Fatalf("reported success before commit: %+v %v", out, err)
	}
	assertLegacyCount(t, db, "household_entry", 0)
	assertLegacyCount(t, db, "household_sub_category", 0)
}

func TestHouseholdBulkSharingConcurrentRetry(t *testing.T) {
	db, s, family := householdReferenceFixture(t)
	original, err := NewTransactionStore(db).CreateV1Transaction(ptr(v1Row()))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	outcomes := make(chan string, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := s.BulkShare(context.Background(), "2", family, h.BulkSharingInput{TransactionIDs: []string{original.TransactionId}}, true)
			if err != nil {
				outcomes <- err.Error()
				return
			}
			outcomes <- out.Results[0].Outcome
		}()
	}
	wg.Wait()
	close(outcomes)
	counts := map[string]int{}
	for outcome := range outcomes {
		counts[outcome]++
	}
	if counts["shared"] != 1 || counts["already-shared"] != 1 {
		t.Fatal(counts)
	}
	assertLegacyCount(t, db, "household_entry", 1)
}

func assertBulkOutcomes(t *testing.T, out *h.BulkSharingResponse, outcomes ...string) {
	t.Helper()
	if out == nil || len(out.Results) != len(outcomes) {
		t.Fatalf("results: %+v", out)
	}
	for i, want := range outcomes {
		if out.Results[i].Outcome != want {
			t.Fatalf("item %d: %+v want %s", i, out.Results[i], want)
		}
	}
}
