//go:build integration

package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"fmt"
	"sync"
	"testing"

	"gorm.io/gorm"
)

func householdReferenceFixture(t *testing.T) (*gorm.DB, *HouseholdStore, string) {
	t.Helper()
	db := legacyTestDB(t)
	s := NewHouseholdStore(db, "reference-test-secret-32-characters")
	f, err := s.Create(context.Background(), "2", h.FamilyInput{Name: "家計", DisplayName: "本人"}, "family-create")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO payment_resource (payment_id,user_no,payment_name,payment_type_id,payment_date,closing_date) VALUES (10,2,'カード',2,27,15),(11,2,'現金',1,NULL,NULL),(12,3,'カード',2,27,15)`).Error; err != nil {
		t.Fatal(err)
	}
	return db, s, stringID(f.ID)
}

func householdReferenceInput() h.EntryInput {
	return h.EntryInput{Transaction: h.TransactionInput{Name: "昼食", Date: "2026-09-29", Amount: 1200, Sign: -1, CategoryID: "1", SubCategoryName: "外食", PaymentID: ptr("10")}}
}

func TestHouseholdAutomaticReferencesCreateAndShare(t *testing.T) {
	db, s, fid := householdReferenceFixture(t)
	ctx := context.Background()
	in := householdReferenceInput()
	first, err := s.CreateEntry(ctx, "2", fid, in, true, "first-entry")
	if err != nil {
		t.Fatal(err)
	}
	if first.HouseholdSubCategoryName == nil || *first.HouseholdSubCategoryName != "外食" || first.HouseholdPaymentName == nil || *first.HouseholdPaymentName != "カード" {
		t.Fatalf("missing automatic references: %+v", first)
	}
	if first.PaymentID != nil || first.SubCategoryID != "" || first.SubCategoryName != "" {
		t.Fatal("personal IDs exposed")
	}
	personal, err := NewTransactionStore(db).GetV1Transaction("2", *first.SourceTransactionID)
	if err != nil {
		t.Fatal(err)
	}
	if personal.PaymentId == nil || *personal.PaymentId != "10" || personal.SubCategoryName != "外食" {
		t.Fatal("personal selections lost")
	}
	payments, err := s.References(ctx, "2", fid, "payments")
	if err != nil {
		t.Fatal(err)
	}
	if len(payments) != 1 || payments[0].PaymentTypeID == nil || *payments[0].PaymentTypeID != "2" || payments[0].PaymentDate == nil || *payments[0].PaymentDate != 27 || payments[0].ClosingDate == nil || *payments[0].ClosingDate != 15 {
		t.Fatalf("payment details not copied: %+v", payments)
	}
	retry, err := s.CreateEntry(ctx, "2", fid, in, true, "first-entry")
	if err != nil || retry.ID != first.ID {
		t.Fatalf("retry: %+v %v", retry, err)
	}
	second, err := s.CreateEntry(ctx, "2", fid, in, true, "second-entry")
	if err != nil {
		t.Fatal(err)
	}
	if *second.HouseholdSubCategoryID != *first.HouseholdSubCategoryID || *second.HouseholdPaymentID != *first.HouseholdPaymentID {
		t.Fatal("identical references duplicated")
	}
	personalInput := v1Row()
	personalInput.PaymentId = ptr("10")
	original, err := NewTransactionStore(db).CreateV1Transaction(&personalInput)
	if err != nil {
		t.Fatal(err)
	}
	shared, err := s.Share(ctx, "2", fid, original.TransactionId, h.EntryInput{SourceVersion: original.Version})
	if err != nil {
		t.Fatal(err)
	}
	if *shared.HouseholdSubCategoryID != *first.HouseholdSubCategoryID || shared.HouseholdPaymentID == nil || *shared.HouseholdPaymentID != *first.HouseholdPaymentID {
		t.Fatal("existing personal share did not resolve references")
	}
}

func TestHouseholdAutomaticReferenceMatching(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		change               string
		samePayment, sameSub bool
	}{
		{"identical", "", true, true},
		{"payment name", "UPDATE payment_resource SET payment_name='別カード' WHERE payment_id=10", false, true},
		{"payment type", "UPDATE payment_resource SET payment_type_id=3 WHERE payment_id=10", false, true},
		{"payment day", "UPDATE payment_resource SET payment_date=28 WHERE payment_id=10", false, true},
		{"closing day", "UPDATE payment_resource SET closing_date=20 WHERE payment_id=10", false, true},
		{"null payment day", "UPDATE payment_resource SET payment_date=NULL WHERE payment_id=10", false, true},
		{"inactive", "UPDATE household_payment SET active=false", false, true},
		{"inactive subcategory", "UPDATE household_sub_category SET active=false", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, s, fid := householdReferenceFixture(t)
			ctx := context.Background()
			in := householdReferenceInput()
			first, err := s.CreateEntry(ctx, "2", fid, in, true, "entry-first")
			if err != nil {
				t.Fatal(err)
			}
			if tc.change != "" {
				if err := db.Exec(tc.change).Error; err != nil {
					t.Fatal(err)
				}
			}
			second, err := s.CreateEntry(ctx, "2", fid, in, true, "entry-second")
			if err != nil {
				t.Fatal(err)
			}
			if (*first.HouseholdPaymentID == *second.HouseholdPaymentID) != tc.samePayment || (*first.HouseholdSubCategoryID == *second.HouseholdSubCategoryID) != tc.sameSub {
				t.Fatalf("wrong match: first=%+v second=%+v", first, second)
			}
		})
	}
}

func TestHouseholdAutomaticReferencesUpdateOnlyChangedSelections(t *testing.T) {
	db, s, fid := householdReferenceFixture(t)
	ctx := context.Background()
	entry, err := s.CreateEntry(ctx, "2", fid, householdReferenceInput(), true, "own-entry")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("UPDATE payment_resource SET payment_name='変更済み' WHERE payment_id=10").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("UPDATE sub_category SET sub_category_name='変更済み'").Error; err != nil {
		t.Fatal(err)
	}
	ts := NewTransactionStore(db)
	original, err := ts.GetV1Transaction("2", *entry.SourceTransactionID)
	if err != nil {
		t.Fatal(err)
	}
	update := v1Row()
	update.TransactionId, update.SubCategoryId, update.SubCategoryName = original.TransactionId, original.SubCategoryId, ""
	update.PaymentId = ptr("10")
	update.Amount = 2000
	if _, _, err := ts.UpdateV1Transaction(&update); err != nil {
		t.Fatal(err)
	}
	unchanged, err := s.Entry(ctx, "2", fid, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if *unchanged.HouseholdPaymentName != "カード" || *unchanged.HouseholdSubCategoryName != "外食" {
		t.Fatal("settings edit rewrote previous family references")
	}
	if err := db.Exec("INSERT INTO sub_category (sub_category_id,user_no,category_id,sub_category_name) VALUES (100,2,2,'外食')").Error; err != nil {
		t.Fatal(err)
	}
	update.CategoryId, update.SubCategoryId, update.PaymentId = "2", "100", ptr("11")
	if _, _, err := ts.UpdateV1Transaction(&update); err != nil {
		t.Fatal(err)
	}
	changed, err := s.Entry(ctx, "2", fid, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if *changed.HouseholdSubCategoryID == *entry.HouseholdSubCategoryID || *changed.HouseholdSubCategoryName != "外食" || *changed.HouseholdPaymentName != "現金" {
		t.Fatal("changed selections not resolved by category and payment")
	}
	update.PaymentId = nil
	if _, _, err := ts.UpdateV1Transaction(&update); err != nil {
		t.Fatal(err)
	}
	cleared, err := s.Entry(ctx, "2", fid, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cleared.HouseholdPaymentID != nil {
		t.Fatal("payment not cleared")
	}
	// Legacy unclassified entries are not backfilled by an amount-only edit.
	if err := db.Table("household_entry").Where("entry_id=?", entry.ID).Update("sub_category_id", nil).Error; err != nil {
		t.Fatal(err)
	}
	update.Amount++
	if _, _, err := ts.UpdateV1Transaction(&update); err != nil {
		t.Fatal(err)
	}
	legacy, err := s.Entry(ctx, "2", fid, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if legacy.HouseholdSubCategoryID != nil {
		t.Fatal("legacy entry backfilled")
	}
}

func TestHouseholdAutomaticReferencesRollback(t *testing.T) {
	db, s, fid := householdReferenceFixture(t)
	if err := db.Exec("ALTER TABLE api_idempotency ADD CONSTRAINT reject_created CHECK (operation NOT LIKE 'own:%')").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateEntry(context.Background(), "2", fid, householdReferenceInput(), true, "rollback-entry"); err == nil {
		t.Fatal("expected idempotency write failure")
	}
	for _, table := range []string{"transaction", "sub_category", "household_entry", "household_sub_category", "household_payment"} {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s left %d rows after rollback", table, count)
		}
	}
}

func TestHouseholdAutomaticReferencesConcurrentMembers(t *testing.T) {
	db, s, fid := householdReferenceFixture(t)
	ctx := context.Background()
	invite, err := s.Issue(ctx, "2", fid, "", "invite-member")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Accept(ctx, "3", "127.0.0.1", h.Credential{Code: invite.Code}, "accept-member"); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			in := householdReferenceInput()
			user := "2"
			if i%2 == 1 {
				user = "3"
				in.Transaction.PaymentID = ptr("12")
			}
			_, err := s.CreateEntry(ctx, user, fid, in, true, fmt.Sprintf("concurrent-%d", i))
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, table := range []string{"household_payment", "household_sub_category"} {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("%s duplicated: %d", table, count)
		}
	}
}

func TestHouseholdAutomaticReferencesReuseManualAndNullableDefinitions(t *testing.T) {
	db, s, fid := householdReferenceFixture(t)
	ctx := context.Background()
	other, err := s.Create(ctx, "4", h.FamilyInput{Name: "別家族", DisplayName: "別本人"}, "other-family")
	if err != nil {
		t.Fatal(err)
	}
	var expectedPayment, expectedSub string
	for i, family := range []string{stringID(other.ID), fid, fid} {
		user := "2"
		if i == 0 {
			user = "4"
		}
		payment, err := s.SaveReference(ctx, user, family, "payments", "", h.ReferenceInput{Name: "現金", PaymentTypeID: ptr("1"), ClosingDate: ptr(31)})
		if err != nil {
			t.Fatal(err)
		}
		sub, err := s.SaveReference(ctx, user, family, "subcategories", "", h.ReferenceInput{Name: "外食", CategoryID: ptr("1")})
		if err != nil {
			t.Fatal(err)
		}
		if i == 1 {
			expectedPayment, expectedSub = payment.ID, sub.ID
		}
	}
	in := householdReferenceInput()
	in.Transaction.PaymentID = ptr("11") // Personal NULL closing date is effectively month-end.
	if err := db.Exec("UPDATE payment_resource SET payment_date=0,closing_date=0 WHERE payment_id=11").Error; err != nil {
		t.Fatal(err)
	}
	entry, err := s.CreateEntry(ctx, "2", fid, in, true, "manual-match")
	if err != nil {
		t.Fatal(err)
	}
	if *entry.HouseholdPaymentID != expectedPayment || *entry.HouseholdSubCategoryID != expectedSub {
		t.Fatal("did not select oldest active match in the same family")
	}
	var count int64
	if err := db.Table("household_payment").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatal("nullable matching created an extra payment")
	}
}

func TestHouseholdAutomaticReferenceUpdateRollback(t *testing.T) {
	db, s, fid := householdReferenceFixture(t)
	ctx := context.Background()
	entry, err := s.CreateEntry(ctx, "2", fid, householdReferenceInput(), true, "update-first")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO sub_category (sub_category_id,user_no,category_id,sub_category_name) VALUES (100,2,2,'別分類')").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("ALTER TABLE household_entry ADD CONSTRAINT reject_update CHECK (version = 1)").Error; err != nil {
		t.Fatal(err)
	}
	update := v1Row()
	update.TransactionId, update.CategoryId, update.SubCategoryId, update.SubCategoryName = *entry.SourceTransactionID, "2", "100", ""
	update.PaymentId = ptr("11")
	update.Amount = 9999
	if _, _, err := NewTransactionStore(db).UpdateV1Transaction(&update); err == nil {
		t.Fatal("expected update failure")
	}
	original, err := NewTransactionStore(db).GetV1Transaction("2", *entry.SourceTransactionID)
	if err != nil {
		t.Fatal(err)
	}
	if original.Amount != 1200 || original.CategoryId != "1" || *original.PaymentId != "10" {
		t.Fatal("personal update was not rolled back")
	}
	after, err := s.Entry(ctx, "2", fid, entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.Version != entry.Version || *after.HouseholdSubCategoryID != *entry.HouseholdSubCategoryID || *after.HouseholdPaymentID != *entry.HouseholdPaymentID {
		t.Fatal("family update was not rolled back")
	}
	for _, table := range []string{"household_payment", "household_sub_category"} {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("%s retained new reference after failed update", table)
		}
	}
}
