//go:build integration

package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestHouseholdShareVersionsArchiveAndPrivacy(t *testing.T) {
	db := legacyTestDB(t)
	ctx := context.Background()
	s := NewHouseholdStore(db, strings.Repeat("q", 32))
	ts := NewTransactionStore(db)
	must := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	f, e := s.Create(ctx, "2", h.FamilyInput{Name: "家族", DisplayName: "本人"}, "create-key")
	must(e)
	fid := stringID(f.ID)
	in := v1Row()
	original, e := ts.CreateV1Transaction(&in)
	must(e)
	status, e := s.ShareStatus(ctx, "2", fid, original.TransactionId)
	must(e)
	if status.State != "none" {
		t.Fatal(status)
	}
	entry, e := s.Share(ctx, "2", fid, original.TransactionId, h.EntryInput{SourceVersion: original.Version})
	must(e)
	must(s.Unshare(ctx, "2", fid, original.TransactionId, entry.Version))
	status, e = s.ShareStatus(ctx, "2", fid, original.TransactionId)
	must(e)
	if status.State != "withdrawn" || status.Version != entry.Version+1 {
		t.Fatal(status)
	}
	if _, e = s.Share(ctx, "2", fid, original.TransactionId, h.EntryInput{SourceVersion: original.Version, Version: h.Version{ExpectedVersion: entry.Version}}); !errors.Is(e, h.Conflict) {
		t.Fatalf("stale reshare: %v", e)
	}
	entry, e = s.Share(ctx, "2", fid, original.TransactionId, h.EntryInput{SourceVersion: original.Version, Version: h.Version{ExpectedVersion: status.Version}})
	must(e)
	rows, _, e := ts.ListV1Transactions("2", "2026-09-01", "shared", "")
	must(e)
	if len(rows) != 1 {
		t.Fatal("shared filter")
	}
	rows, _, e = ts.ListV1Transactions("2", "2026-09-01", "private", "")
	must(e)
	if len(rows) != 0 {
		t.Fatal("private filter")
	}
	f, e = s.Get(ctx, "2", fid)
	must(e)
	must(db.Table("users").Where("user_no=2").Update("default_transaction_scope", "household").Error)
	must(s.Leave(ctx, "2", fid, "", f.Version, true))
	snapshot, e := s.Entry(ctx, "2", fid, entry.ID)
	must(e)
	if snapshot.Kind != "snapshot" || snapshot.Permissions.Correct {
		t.Fatal("archived write permissions")
	}
	if _, e = s.CreateEntry(ctx, "2", fid, h.EntryInput{Transaction: h.TransactionInput{Name: "無効", Date: "2026-09-29", Amount: 1, Sign: -1, CategoryID: "1"}, Payer: h.Payer{Kind: "common"}}, false, "archived-key"); !errors.Is(e, h.Forbidden) {
		t.Fatalf("archived write: %v", e)
	}
	var scope string
	must(db.Table("users").Select("default_transaction_scope").Where("user_no=2").Scan(&scope).Error)
	if scope != "personal" {
		t.Fatal(scope)
	}
	_, e = s.Create(ctx, "2", h.FamilyInput{Name: "次の家族", DisplayName: "本人"}, "another-create")
	must(e)
	must(ts.DeleteV1Transaction("2", original.TransactionId))
	rows, _, e = ts.ListV1Transactions("2", "2026-09-01", "all", "")
	must(e)
	if len(rows) != 0 {
		t.Fatal("deleted original in list")
	}
	timeline, e := ts.GetTimelineData("2", "2026-09-01")
	must(e)
	if len(*timeline) != 0 {
		t.Fatal("deleted original in legacy timeline")
	}
	snapshot, e = s.Entry(ctx, "2", fid, entry.ID)
	must(e)
	if snapshot.Amount != 1200 {
		t.Fatal("archive snapshot changed")
	}
	encoded, e := json.Marshal(snapshot)
	must(e)
	for _, private := range []string{"user_no", "source_transaction_id", "sub_category_name", "code_digest", "token_digest"} {
		if strings.Contains(string(encoded), `"`+private+`"`) {
			t.Fatalf("private field %s", private)
		}
	}
}

func TestHouseholdInvitationRevocationAndRejoin(t *testing.T) {
	db := legacyTestDB(t)
	ctx := context.Background()
	s := NewHouseholdStore(db, strings.Repeat("r", 32))
	must := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	f, e := s.Create(ctx, "2", h.FamilyInput{Name: "家族", DisplayName: "本人"}, "create-key")
	must(e)
	fid := stringID(f.ID)
	invite, e := s.Issue(ctx, "2", fid, "", "invite-key")
	must(e)
	_, e = s.Preview(ctx, "3", "first-ip", h.Credential{Code: invite.Code})
	must(e)
	replacement, e := s.Issue(ctx, "2", fid, stringID(invite.ID), "replace-key")
	must(e)
	if _, e = s.Accept(ctx, "3", "first-ip", h.Credential{Token: invite.Token}, "old-accept"); e == nil {
		t.Fatal("revoked token accepted")
	}
	_, e = s.Accept(ctx, "3", "first-ip", h.Credential{Token: replacement.Token}, "new-accept")
	must(e)
	in := h.EntryInput{Transaction: h.TransactionInput{Name: "昼食", Date: "2026-09-29", Amount: 1200, Sign: -1, CategoryID: "1", SubCategoryName: "外食"}}
	entry, e := s.CreateEntry(ctx, "3", fid, in, true, "entry-key")
	must(e)
	f, e = s.Get(ctx, "2", fid)
	must(e)
	must(s.Leave(ctx, "3", fid, "", f.Version, false))
	rejoin, e := s.Issue(ctx, "2", fid, "", "rejoin-key")
	must(e)
	_, e = s.Accept(ctx, "3", "first-ip", h.Credential{Code: rejoin.Code}, "rejoin-accept")
	must(e)
	if _, e = s.Share(ctx, "3", fid, *entry.SourceTransactionID, h.EntryInput{SourceVersion: *entry.SourceVersion}); e == nil {
		t.Fatal("snapshot reconnected")
	}
	members, e := s.Members(ctx, "2", fid)
	must(e)
	var next string
	for _, m := range members {
		if m.UserNo == 3 {
			next = stringID(m.ID)
		}
	}
	pending, e := s.Issue(ctx, "2", fid, "", "pending-key")
	must(e)
	f, e = s.Get(ctx, "2", fid)
	must(e)
	must(s.Transfer(ctx, "2", fid, h.MemberInput{MemberID: next, Version: h.Version{ExpectedVersion: f.Version}}))
	if _, e = s.Accept(ctx, "4", "second-ip", h.Credential{Code: pending.Code}, "after-transfer"); e == nil {
		t.Fatal("old admin invite usable")
	}
	expired, e := s.Issue(ctx, "3", fid, "", "expired-key")
	must(e)
	must(db.Model(&h.Invitation{}).Where("invitation_id=?", expired.ID).Update("expires_at", "2000-01-01").Error)
	if _, e = s.Accept(ctx, "4", "third-ip", h.Credential{Code: expired.Code}, "after-expired"); e == nil {
		t.Fatal("expired invite usable")
	}
	for i := 0; i < 21; i++ {
		_, e = s.Preview(ctx, "5", "guesses-ip", h.Credential{Code: "0000000000"})
	}
	var domain *h.Error
	if !errors.As(e, &domain) || domain.Code != "RATE_LIMITED" {
		t.Fatalf("rate limit: %v", e)
	}
}
