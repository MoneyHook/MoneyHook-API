package migration

import "gorm.io/gorm"

// MigrateHouseholds adds family records without changing personal ownership.
func MigrateHouseholds(db *gorm.DB) error {
	statements := []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS default_transaction_scope varchar(16) NOT NULL DEFAULT 'personal'`,
		`ALTER TABLE "transaction" ADD COLUMN IF NOT EXISTS version bigint NOT NULL DEFAULT 1`,
		`ALTER TABLE "transaction" ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`ALTER TABLE "transaction" ADD COLUMN IF NOT EXISTS deleted_at timestamptz`,
		`CREATE TABLE IF NOT EXISTS household (household_id bigserial PRIMARY KEY, name varchar(64) NOT NULL, state varchar(16) NOT NULL CHECK (state IN ('active','archived')), version bigint NOT NULL DEFAULT 1, created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, archived_at timestamptz)`,
		`CREATE TABLE IF NOT EXISTS household_member (member_id bigserial PRIMARY KEY, household_id bigint NOT NULL REFERENCES household, user_no bigint NOT NULL REFERENCES users(user_no), role varchar(16) NOT NULL CHECK(role IN ('admin','member')), state varchar(16) NOT NULL CHECK(state IN ('active','left','archived')), slot_no smallint, display_name varchar(32) NOT NULL, joined_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, left_at timestamptz, UNIQUE(household_id,user_no), UNIQUE(household_id,member_id), CHECK((state='active' AND slot_no BETWEEN 1 AND 3) OR (state<>'active' AND slot_no IS NULL)), CHECK(state<>'active' OR slot_no IS NOT NULL))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_household_active_user ON household_member(user_no) WHERE state='active'`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_household_slot ON household_member(household_id,slot_no) WHERE state='active'`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_household_admin ON household_member(household_id) WHERE state='active' AND role='admin'`,
		`CREATE TABLE IF NOT EXISTS household_invitation (invitation_id bigserial PRIMARY KEY, household_id bigint NOT NULL REFERENCES household, issued_by bigint NOT NULL REFERENCES household_member(member_id), code_digest varchar(64) NOT NULL UNIQUE, token_digest varchar(64) NOT NULL UNIQUE, expires_at timestamptz NOT NULL, consumed_by bigint REFERENCES users(user_no), consumed_at timestamptz, revoked_at timestamptz, created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS household_payment (payment_id bigserial PRIMARY KEY, household_id bigint NOT NULL REFERENCES household, name varchar(32) NOT NULL, payment_type_id bigint NOT NULL REFERENCES payment_type, payment_date smallint CHECK(payment_date BETWEEN 1 AND 31), closing_date smallint CHECK(closing_date BETWEEN 1 AND 31), active boolean NOT NULL DEFAULT true, version bigint NOT NULL DEFAULT 1, UNIQUE(household_id,payment_id))`,
		`CREATE TABLE IF NOT EXISTS household_sub_category (sub_category_id bigserial PRIMARY KEY, household_id bigint NOT NULL REFERENCES household, category_id bigint NOT NULL REFERENCES category, name varchar(16) NOT NULL, active boolean NOT NULL DEFAULT true, version bigint NOT NULL DEFAULT 1, UNIQUE(household_id,sub_category_id))`,
		`CREATE TABLE IF NOT EXISTS household_entry (entry_id bigserial PRIMARY KEY, household_id bigint NOT NULL REFERENCES household, kind varchar(16) NOT NULL CHECK(kind IN ('shared','proxy','snapshot')), source_transaction_id bigint REFERENCES "transaction"(transaction_id), payer_member_id bigint, payment_id bigint, sub_category_id bigint, state varchar(16) NOT NULL DEFAULT 'active' CHECK(state IN ('active','withdrawn','deleted')), version bigint NOT NULL DEFAULT 1, created_by bigint NOT NULL REFERENCES household_member(member_id), updated_by bigint NOT NULL REFERENCES household_member(member_id), created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY(household_id,payer_member_id) REFERENCES household_member(household_id,member_id), FOREIGN KEY(household_id,payment_id) REFERENCES household_payment(household_id,payment_id), FOREIGN KEY(household_id,sub_category_id) REFERENCES household_sub_category(household_id,sub_category_id), CHECK((kind='proxy' AND source_transaction_id IS NULL) OR (kind='shared' AND source_transaction_id IS NOT NULL AND payer_member_id IS NOT NULL) OR (kind='snapshot' AND source_transaction_id IS NOT NULL)), UNIQUE(household_id,source_transaction_id))`,
		`CREATE TABLE IF NOT EXISTS household_entry_data (entry_id bigint PRIMARY KEY REFERENCES household_entry, payload jsonb NOT NULL, corrected_payload jsonb, captured_at timestamptz, source_version bigint)`,
		`ALTER TABLE household_entry_data ADD COLUMN IF NOT EXISTS corrected_payload jsonb`,
		`CREATE TABLE IF NOT EXISTS household_event (event_id bigserial PRIMARY KEY, household_id bigint NOT NULL REFERENCES household, actor bigint NOT NULL REFERENCES users(user_no), target_id bigint, action varchar(32) NOT NULL, created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`CREATE TABLE IF NOT EXISTS api_idempotency (user_no bigint NOT NULL REFERENCES users(user_no), operation varchar(128) NOT NULL, key varchar(128) NOT NULL, request_digest varchar(64) NOT NULL, resource_id bigint NOT NULL, created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY(user_no,operation,key))`,
		`CREATE TABLE IF NOT EXISTS household_invitation_attempt (bucket varchar(80) PRIMARY KEY, attempts integer NOT NULL, expires_at timestamptz NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS ix_household_entry_family ON household_entry(household_id,state,entry_id)`,
		`CREATE INDEX IF NOT EXISTS ix_household_entry_source ON household_entry(source_transaction_id)`,
		`CREATE INDEX IF NOT EXISTS ix_household_invitation_expiry ON household_invitation(household_id,expires_at)`,
	}
	for _, s := range statements {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return removeHouseholdEntryHistory(db)
}
