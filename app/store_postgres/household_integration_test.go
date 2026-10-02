//go:build integration

package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestHouseholdOwnershipAndDeparture(t *testing.T) {
	db := legacyTestDB(t)
	s := NewHouseholdStore(db, strings.Repeat("x", 32))
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	f, err := s.Create(ctx, "2", h.FamilyInput{Name: "家計", DisplayName: "親"}, "create-key")
	must(err)
	fid := stringID(f.ID)
	invite, err := s.Issue(ctx, "2", fid, "", "invite-key")
	must(err)
	_, err = s.Accept(ctx, "3", "127.0.0.1", h.Credential{Code: invite.Code}, "accept-key")
	must(err)
	// One credential can only be consumed once, whether using the code or URL.
	if _, err = s.Accept(ctx, "4", "127.0.0.1", h.Credential{Token: invite.Token}, "accept-other"); err == nil {
		t.Fatal("reused invitation")
	}
	in := h.EntryInput{Transaction: h.TransactionInput{Name: "昼食", Date: "2026-09-29", Amount: 1200, Sign: -1, CategoryID: "1", SubCategoryName: "外食"}}
	own, err := s.CreateEntry(ctx, "3", fid, in, true, "own-entry-key")
	must(err)
	retry, err := s.CreateEntry(ctx, "3", fid, in, true, "own-entry-key")
	must(err)
	if own.ID != retry.ID {
		t.Fatal("idempotency duplicated original")
	}
	other, err := s.Entry(ctx, "2", fid, own.ID)
	must(err)
	if other.SourceTransactionID != nil || other.Permissions.Edit || other.Permissions.Delete {
		t.Fatal("owner information/permissions leaked")
	}
	if err = s.Unshare(ctx, "2", fid, *own.SourceTransactionID, own.Version); !errors.Is(err, h.Forbidden) {
		t.Fatalf("nonowner unshare: %v", err)
	}
	if _, err = s.UpdateProxy(ctx, "2", fid, own.ID, in); err == nil {
		t.Fatal("original mutated as proxy")
	}
	members, err := s.Members(ctx, "2", fid)
	must(err)
	var departing string
	for _, m := range members {
		if m.UserNo == 3 {
			departing = stringID(m.ID)
		}
	}
	proxyIn := in
	proxyIn.Payer = h.Payer{Kind: "member", MemberID: &departing}
	proxy, err := s.CreateEntry(ctx, "2", fid, proxyIn, false, "proxy-entry-key")
	must(err)
	var count int64
	must(db.Table("transaction").Count(&count).Error)
	if count != 1 {
		t.Fatalf("proxy wrote personal pool: %d", count)
	}
	proxyIn.ExpectedVersion = proxy.Version
	proxyIn.Transaction.Amount = 2000
	_, err = s.UpdateProxy(ctx, "3", fid, proxy.ID, proxyIn)
	must(err)
	refs, err := s.SaveReference(ctx, "2", fid, "payments", "", h.ReferenceInput{Name: "家族現金", PaymentTypeID: ptr("1")})
	must(err)
	if refs.ID == "" {
		t.Fatal("reference id missing")
	}
	// Personal changes synchronize while sharing, and stale writes fail.
	ts := NewTransactionStore(db)
	update := v1Row()
	update.UserId = "3"
	update.TransactionId = *own.SourceTransactionID
	source, err := ts.GetV1Transaction("3", *own.SourceTransactionID)
	must(err)
	update.SubCategoryName = ""
	update.SubCategoryId = source.SubCategoryId
	update.TransactionName = "修正昼食"
	update.Amount = 1500
	update.ExpectedVersion = own.SourceVersion
	_, _, err = ts.UpdateV1Transaction(&update)
	must(err)
	if _, _, err = ts.UpdateV1Transaction(&update); err == nil {
		t.Fatal("stale original update succeeded")
	}
	current, err := s.Entry(ctx, "2", fid, own.ID)
	must(err)
	if current.Amount != 1500 {
		t.Fatal("original update not projected")
	}
	f, err = s.Get(ctx, "2", fid)
	must(err)
	must(s.Leave(ctx, "3", fid, "", f.Version, false))
	snapshot, err := s.Entry(ctx, "2", fid, own.ID)
	must(err)
	if snapshot.Kind != "snapshot" || snapshot.Permissions.Edit || !snapshot.Permissions.Correct {
		t.Fatalf("bad snapshot: %+v", snapshot)
	}
	if _, err = s.Entry(ctx, "3", fid, own.ID); !errors.Is(err, h.NotFound) {
		t.Fatalf("departed read: %v", err)
	}
	original, err := ts.GetV1Transaction("3", *own.SourceTransactionID)
	must(err)
	update.ExpectedVersion = &original.Version
	update.Amount = 9000
	_, _, err = ts.UpdateV1Transaction(&update)
	must(err)
	snapshot, err = s.Entry(ctx, "2", fid, own.ID)
	must(err)
	if snapshot.Amount != 1500 {
		t.Fatal("snapshot followed departed original")
	}
	var capturedPayload string
	must(db.Table("household_entry_data").Select("payload").Where("entry_id = ?", own.ID).Scan(&capturedPayload).Error)
	correction := proxyIn
	correction.ExpectedVersion = snapshot.Version
	correction.Transaction.Amount = 1800
	corrected, err := s.Correct(ctx, "2", fid, own.ID, correction)
	must(err)
	if corrected.Amount != 1800 || !corrected.Corrected {
		t.Fatal("correction missing")
	}
	correction.ExpectedVersion = corrected.Version
	correction.Excluded = true
	correction.Payer = h.Payer{Kind: "common"}
	corrected, err = s.Correct(ctx, "2", fid, own.ID, correction)
	must(err)
	if !corrected.Excluded || corrected.Payer.Kind != "common" {
		t.Fatal("snapshot exclusion/corrected payer missing")
	}
	var unchangedPayload string
	must(db.Table("household_entry_data").Select("payload").Where("entry_id = ?", own.ID).Scan(&unchangedPayload).Error)
	if capturedPayload != unchangedPayload {
		t.Fatal("correction overwrote immutable snapshot data")
	}
	page, err := s.Entries(ctx, "2", fid, h.Filter{Month: "2026-09-01"})
	must(err)
	if len(page.Entries) != 2 {
		t.Fatalf("entries: %d", len(page.Entries))
	}
	must(ts.DeleteV1Transaction("3", *own.SourceTransactionID))
	snapshot, err = s.Entry(ctx, "2", fid, own.ID)
	must(err)
	if snapshot.Amount != 1800 {
		t.Fatal("original delete changed snapshot")
	}
}

func TestHouseholdConcurrentCapacityAndInvitation(t *testing.T) {
	db := legacyTestDB(t)
	s := NewHouseholdStore(db, strings.Repeat("z", 32))
	ctx := context.Background()
	f, err := s.Create(ctx, "2", h.FamilyInput{Name: "家計", DisplayName: "親"}, "create-key")
	if err != nil {
		t.Fatal(err)
	}
	fid := stringID(f.ID)
	a, err := s.Issue(ctx, "2", fid, "", "invite-one")
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.Issue(ctx, "2", fid, "", "invite-two")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Issue(ctx, "2", fid, "", "invite-three"); err == nil {
		t.Fatal("reserved more than 3 slots")
	}
	var wg sync.WaitGroup
	results := make(chan error, 3)
	for i, u := range []string{"3", "4", "5"} {
		wg.Add(1)
		go func(i int, u string) {
			defer wg.Done()
			code := a.Code
			if i == 2 {
				code = b.Code
			}
			_, e := s.Accept(ctx, u, "127.0.0.2", h.Credential{Code: code}, fmt.Sprintf("accept-key-%s", u))
			results <- e
		}(i, u)
	}
	wg.Wait()
	close(results)
	success := 0
	for e := range results {
		if e == nil {
			success++
		}
	}
	if success != 2 {
		t.Fatalf("accepted %d, expected two", success)
	}
	var count int64
	db.Table("household_member").Where("state='active'").Count(&count)
	if count != 3 {
		t.Fatalf("active count %d", count)
	}
}
