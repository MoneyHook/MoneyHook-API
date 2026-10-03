//go:build integration

package store_postgres

import (
	"MoneyHook/MoneyHook-API/db/migration"
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"reflect"
	"testing"
)

func TestHouseholdMigrationRemovesEvents(t *testing.T) {
	db, store, familyID := householdReferenceFixture(t)
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	if db.Migrator().HasTable("household_event") {
		t.Fatal("fresh schema contains event history")
	}
	family, err := store.Get(ctx, "2", familyID)
	must(err)
	entry, err := store.CreateEntry(ctx, "2", familyID, householdReferenceInput(), true, "migration-event-entry")
	must(err)
	must(db.Exec(`CREATE TABLE household_event (
		event_id bigserial PRIMARY KEY,
		household_id bigint NOT NULL REFERENCES household,
		actor bigint NOT NULL REFERENCES users(user_no), target_id bigint,
		action varchar(32) NOT NULL,
		created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP)`).Error)
	must(db.Exec("INSERT INTO household_event (household_id,actor,target_id,action) VALUES (?,2,?,'created')", familyID, family.MemberID).Error)
	for range 2 {
		must(migration.MigrateHouseholds(db))
		if db.Migrator().HasTable("household_event") {
			t.Fatal("migration retained event history")
		}
	}
	currentFamily, err := store.Get(ctx, "2", familyID)
	must(err)
	if !reflect.DeepEqual(family, currentFamily) {
		t.Fatalf("migration changed family or membership: %+v", currentFamily)
	}
	currentEntry, err := store.Entry(ctx, "2", familyID, entry.ID)
	must(err)
	if !reflect.DeepEqual(entry, currentEntry) {
		t.Fatalf("migration changed shared transaction: %+v", currentEntry)
	}
	retry, err := store.CreateEntry(ctx, "2", familyID, householdReferenceInput(), true, "migration-event-entry")
	must(err)
	if retry.ID != entry.ID {
		t.Fatal("migration lost transaction idempotency")
	}
}

func TestHouseholdMigrationPreservesLatestCorrection(t *testing.T) {
	db, store, familyID := householdReferenceFixture(t)
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	if db.Migrator().HasTable("household_entry_revision") {
		t.Fatal("fresh schema contains obsolete table")
	}
	if db.Migrator().HasColumn("household_entry_data", "source_version") {
		t.Fatal("fresh schema contains obsolete snapshot source version")
	}
	entry, err := store.CreateEntry(ctx, "2", familyID, householdReferenceInput(), true, "migration-entry")
	must(err)
	// Reproduce an existing snapshot and its old correction storage.
	must(db.Exec("UPDATE household_entry SET kind='snapshot', version=5 WHERE entry_id=?", entry.ID).Error)
	must(db.Exec("ALTER TABLE household_entry_data ADD COLUMN source_version bigint").Error)
	must(db.Exec("INSERT INTO household_entry_data (entry_id,payload,captured_at,source_version) VALUES (?,?::jsonb,CURRENT_TIMESTAMP,1)", entry.ID, string(publicEntry(entry))).Error)
	must(db.Exec("ALTER TABLE household_entry_data DROP COLUMN corrected_payload").Error)
	must(db.Exec(`CREATE TABLE household_entry_revision (
		entry_id bigint NOT NULL, entry_version bigint NOT NULL,
		action text NOT NULL, after_data jsonb NOT NULL,
		UNIQUE(entry_id,entry_version))`).Error)
	for _, row := range []struct {
		version int
		action  string
		amount  int64
	}{{4, "corrected", 2400}, {3, "corrected", 1800}, {5, "captured", 1200}} {
		value := *entry
		value.Amount, value.SignedAmount = row.amount, -row.amount
		value.Date = "2026-10-01"
		value.Excluded = true
		value.Payer = h.Payer{Kind: "common", DisplayName: "家族共通"}
		must(db.Exec("INSERT INTO household_entry_revision VALUES (?,?,?,?::jsonb)", entry.ID, row.version, row.action, string(publicEntry(&value))).Error)
	}
	// Failed data transfer must not delete the source data.
	must(db.Exec("ALTER TABLE household_entry_data ADD COLUMN corrected_payload jsonb CHECK (corrected_payload IS NULL)").Error)
	if err := migration.MigrateHouseholds(db); err == nil {
		t.Fatal("expected correction transfer to fail")
	}
	if !db.Migrator().HasTable("household_entry_revision") {
		t.Fatal("failed migration deleted source data")
	}
	must(db.Exec("ALTER TABLE household_entry_data DROP COLUMN corrected_payload").Error)
	must(migration.MigrateHouseholds(db))
	must(migration.MigrateHouseholds(db))
	if db.Migrator().HasTable("household_entry_revision") {
		t.Fatal("migration retained obsolete table")
	}
	if db.Migrator().HasColumn("household_entry_data", "source_version") {
		t.Fatal("migration retained obsolete snapshot source version")
	}
	current, err := store.Entry(ctx, "2", familyID, entry.ID)
	must(err)
	if current.Amount != 2400 || current.Date != "2026-10-01" || !current.Corrected || !current.Excluded || current.Payer.Kind != "common" || current.CapturedAt == nil || current.Version != 5 {
		t.Fatalf("latest correction was not preserved: %+v", current)
	}
	var originalAmount int64
	must(db.Raw("SELECT (payload->>'amount')::bigint FROM household_entry_data WHERE entry_id=?", entry.ID).Scan(&originalAmount).Error)
	if originalAmount != entry.Amount {
		t.Fatal("migration changed captured payload")
	}
	page, err := store.Entries(ctx, "2", familyID, h.Filter{Month: "2026-10-01"})
	must(err)
	if len(page.Entries) != 1 || page.Entries[0].Amount != 2400 {
		t.Fatal("list did not use corrected date and amount")
	}
	// Further corrections replace the single effective value.
	in := h.EntryInput{Version: h.Version{ExpectedVersion: current.Version}, Transaction: current.TransactionInput, Payer: h.Payer{Kind: "common"}}
	in.Transaction.Amount = 3000
	updated, err := store.Correct(ctx, "2", familyID, entry.ID, in)
	must(err)
	if updated.Amount != 3000 || updated.Excluded || !updated.Corrected {
		t.Fatalf("correction after migration failed: %+v", updated)
	}
	if _, err := store.Correct(ctx, "2", familyID, entry.ID, in); err != h.Conflict {
		t.Fatalf("stale correction: %v", err)
	}
}
