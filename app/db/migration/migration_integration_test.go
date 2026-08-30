//go:build integration

package migration

import (
	userdomain "MoneyHook/MoneyHook-API/user"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"MoneyHook/MoneyHook-API/store_postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestIntegrationPostgreSQL(t *testing.T) {
	rawDSN := os.Getenv("MIGRATION_TEST_POSTGRES_DSN")
	if rawDSN == "" {
		t.Skip("MIGRATION_TEST_POSTGRES_DSN is not set")
	}

	admin, err := gorm.Open(postgres.Open(rawDSN), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to PostgreSQL test server: %v", err)
	}
	schemaName := fmt.Sprintf("moneyhooks_migration_test_%d", time.Now().UnixNano())
	mustExec(t, admin, fmt.Sprintf("CREATE SCHEMA %s", quoteIdentifier(schemaName)))
	t.Cleanup(func() {
		admin.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", quoteIdentifier(schemaName)))
		closeGormDB(admin)
	})

	parsed, err := url.Parse(rawDSN)
	if err != nil {
		t.Fatalf("parse PostgreSQL DSN: %v", err)
	}
	query := parsed.Query()
	query.Set("search_path", schemaName)
	parsed.RawQuery = query.Encode()
	testDB, err := gorm.Open(postgres.New(postgres.Config{DSN: parsed.String(), PreferSimpleProtocol: true}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,
	})
	if err != nil {
		t.Fatalf("connect to PostgreSQL test schema: %v", err)
	}
	t.Cleanup(func() { closeGormDB(testDB) })

	exerciseMigration(t, testDB, schemaName)
}

func exerciseMigration(t *testing.T, db *gorm.DB, databaseName string) {
	t.Helper()
	ctx := context.Background()
	createLegacyAuthSchema(t, db)
	err := Run(ctx, db, databaseName, Options{})
	if err == nil || !strings.Contains(err.Error(), "duplicated") {
		t.Fatalf("duplicate legacy identity migration error = %v", err)
	}
	if db.Migrator().HasTable("user_token") || db.Migrator().HasColumn("users", "email") || db.Migrator().HasColumn("users", "password") {
		t.Fatal("legacy authentication storage was not removed")
	}
	mustExec(t, db, "DELETE FROM users WHERE user_no = ?", 98)
	if err := Run(ctx, db, databaseName, Options{}); err != nil {
		t.Fatalf("migrate empty database: %v", err)
	}
	var preservedLegacyUser int64
	if err := db.Table("users").Where("user_no = ? AND user_id = ?", 99, "duplicate-legacy-id").Count(&preservedLegacyUser).Error; err != nil || preservedLegacyUser != 1 {
		t.Fatalf("legacy user identity was not preserved: count=%d error=%v", preservedLegacyUser, err)
	}
	exerciseFirebaseIdentityResolution(t, db)
	assertCount(t, db, "category", 29)
	assertCount(t, db, "sub_category", 91)
	assertCount(t, db, "payment_type", 3)

	mustExec(t, db, "INSERT INTO users (user_id) VALUES (?)", "integration-sequence-user")
	var generatedUserNo uint64
	if err := db.Table("users").Select("user_no").Where("user_id = ?", "integration-sequence-user").Scan(&generatedUserNo).Error; err != nil {
		t.Fatalf("read generated user number: %v", err)
	}
	if generatedUserNo <= systemUserNo {
		t.Fatalf("generated user number = %d, want greater than %d", generatedUserNo, systemUserNo)
	}

	mustExec(t, db, "UPDATE category SET category_name = ?, order_num = ? WHERE category_id = ?", "broken", 999, 1)
	mustExec(t, db, "DELETE FROM sub_category WHERE user_no = ? AND category_id = ? AND sub_category_name = ?", systemUserNo, 29, "なし")
	mustExec(t, db, "INSERT INTO sub_category (user_no, category_id, sub_category_name) VALUES (?, ?, ?)", systemUserNo, 29, "custom")
	mustExec(t, db, fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", quoteIdentifier("monthly_transaction"), quoteIdentifier("include_flg")))
	mustExec(t, db, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s VARCHAR(32)", quoteIdentifier("payment_type"), quoteIdentifier("preserved_extra_column")))
	mustExec(t, db, fmt.Sprintf("CREATE TABLE %s (%s INTEGER)", quoteIdentifier("preserved_extra_table"), quoteIdentifier("id")))
	paymentTypeIndexes, err := loadUniqueIndexes(ctx, db, "payment_type")
	if err != nil {
		t.Fatalf("load payment type indexes: %v", err)
	}
	paymentTypeIndexName := ""
	for _, index := range paymentTypeIndexes {
		if sameStringSet(index.Columns, []string{"payment_type_name"}) {
			paymentTypeIndexName = index.Name
			break
		}
	}
	if paymentTypeIndexName == "" {
		t.Fatal("payment type unique index was not found before drift test")
	}

	mustExec(t, db, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", quoteIdentifier("payment_type"), quoteIdentifier(paymentTypeIndexName)))
	mustExec(t, db, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", quoteIdentifier("monthly_transaction"), quoteIdentifier("fk_monthly_transaction_user")))

	if err := Run(ctx, db, databaseName, Options{}); err != nil {
		t.Fatalf("repair schema drift: %v", err)
	}
	if !db.Migrator().HasColumn("monthly_transaction", "include_flg") {
		t.Error("missing column was not restored")
	}
	if !db.Migrator().HasColumn("payment_type", "preserved_extra_column") {
		t.Error("extra column was removed")
	}
	if !db.Migrator().HasTable("preserved_extra_table") {
		t.Error("extra table was removed")
	}
	assertCount(t, db, "sub_category", 92)

	var category categorySchema
	if err := db.Where("category_id = ?", 1).Take(&category).Error; err != nil {
		t.Fatalf("load repaired category: %v", err)
	}
	if category.CategoryName != "食費" || category.OrderNum != 1 {
		t.Errorf("category was not reconciled: %+v", category)
	}

	indexes, err := loadUniqueIndexes(ctx, db, "payment_type")
	if err != nil || !containsUniqueIndex(indexes, []string{"payment_type_name"}) {
		t.Errorf("unique index was not restored: indexes=%v error=%v", indexes, err)
	}
	keys, err := loadForeignKeys(ctx, db, "monthly_transaction")
	if err != nil || !containsForeignKey(keys, foreignKeyRequirement{Column: "user_no", ReferencedTable: "users", ReferencedColumn: "user_no"}) {
		t.Errorf("foreign key was not restored: keys=%v error=%v", keys, err)
	}

	if err := Run(ctx, db, databaseName, Options{}); err != nil {
		t.Fatalf("repeat idempotent migration: %v", err)
	}
	assertCount(t, db, "sub_category", 92)

	var waitGroup sync.WaitGroup
	errors := make(chan error, 2)
	for range 2 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			errors <- Run(ctx, db, databaseName, Options{})
		}()
	}
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Errorf("concurrent migration: %v", err)
		}
	}

	mustExec(t, db, "UPDATE users SET user_id = ? WHERE user_no = ?", "conflicting-system-user", systemUserNo)
	err = Run(ctx, db, databaseName, Options{})
	if err == nil || !strings.Contains(err.Error(), "belongs to user_id") {
		t.Fatalf("system user conflict error = %v", err)
	}
	mustExec(t, db, "UPDATE users SET user_id = ? WHERE user_no = ?", systemUserID, systemUserNo)

	exerciseSampleSeed(t, db, databaseName)
	exerciseLegacySchema(t, db, databaseName)
}

func exerciseFirebaseIdentityResolution(t *testing.T, db *gorm.DB) {
	t.Helper()
	var store userdomain.Store = store_postgres.NewUserStore(db)

	legacyDigest := sha256.Sum256([]byte("second@example.com"))
	legacyUserID := hex.EncodeToString(legacyDigest[:])
	mustExec(t, db, "UPDATE users SET user_id = ? WHERE user_no = ?", legacyUserID, 99)
	if err := db.Table("transaction").Create(map[string]any{
		"user_no":            99,
		"transaction_name":   "移行前データ",
		"transaction_amount": -100,
		"transaction_date":   "2026-08-29",
		"category_id":        1,
		"sub_category_id":    1,
		"fixed_flg":          false,
	}).Error; err != nil {
		t.Fatalf("create legacy user's business data: %v", err)
	}

	userNo, err := store.ResolveFirebaseUser("firebase-uid-migrated", legacyUserID)
	if err != nil || userNo != "99" {
		t.Fatalf("migrate legacy email hash: userNo=%q error=%v", userNo, err)
	}
	var preservedTransactions int64
	if err := db.Table("transaction").Where("user_no = ?", 99).Count(&preservedTransactions).Error; err != nil || preservedTransactions != 1 {
		t.Fatalf("business data was not preserved: count=%d error=%v", preservedTransactions, err)
	}
	userNo, err = store.ResolveFirebaseUser("firebase-uid-migrated", legacyUserID)
	if err != nil || userNo != "99" {
		t.Fatalf("resolve existing Firebase UID: userNo=%q error=%v", userNo, err)
	}

	newUserNo, err := store.ResolveFirebaseUser("firebase-uid-new", "missing-legacy-id")
	if err != nil || newUserNo == "" {
		t.Fatalf("create Firebase user: userNo=%q error=%v", newUserNo, err)
	}

	mustExec(t, db, "INSERT INTO users (user_id) VALUES (?)", "firebase-uid-conflict")
	mustExec(t, db, "INSERT INTO users (user_id) VALUES (?)", "legacy-id-conflict")
	if _, err := store.ResolveFirebaseUser("firebase-uid-conflict", "legacy-id-conflict"); !errors.Is(err, userdomain.ErrIdentityConflict) {
		t.Fatalf("identity conflict error = %v", err)
	}

	const concurrentRequests = 4
	resolved := make(chan string, concurrentRequests)
	resolutionErrors := make(chan error, concurrentRequests)
	var waitGroup sync.WaitGroup
	for range concurrentRequests {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			resolvedUserNo, resolveErr := store.ResolveFirebaseUser("firebase-uid-concurrent", "missing-concurrent-legacy-id")
			resolved <- resolvedUserNo
			resolutionErrors <- resolveErr
		}()
	}
	waitGroup.Wait()
	close(resolved)
	close(resolutionErrors)
	for resolveErr := range resolutionErrors {
		if resolveErr != nil {
			t.Fatalf("concurrent first access: %v", resolveErr)
		}
	}
	var expectedUserNo string
	for resolvedUserNo := range resolved {
		if resolvedUserNo == "" {
			t.Fatal("concurrent first access returned an empty user number")
		}
		if expectedUserNo == "" {
			expectedUserNo = resolvedUserNo
		}
		if resolvedUserNo != expectedUserNo {
			t.Fatalf("concurrent first access returned multiple users: %q and %q", expectedUserNo, resolvedUserNo)
		}
	}
}

func exerciseSampleSeed(t *testing.T, db *gorm.DB, databaseName string) {
	t.Helper()
	ctx := context.Background()
	var sampleCount int64
	if err := db.Model(&userSchema{}).Where("user_id = ?", sampleUserID).Count(&sampleCount).Error; err != nil {
		t.Fatalf("count disabled sample seed: %v", err)
	}
	if sampleCount != 0 {
		t.Fatalf("sample user exists while seed is disabled")
	}
	options := Options{
		EnableSeedData:    true,
		SeedReferenceTime: time.Date(2026, time.August, 29, 12, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)),
	}
	if err := Run(ctx, db, databaseName, options); err != nil {
		t.Fatalf("insert sample seed: %v", err)
	}
	var sampleUser userSchema
	if err := db.Where("user_id = ?", sampleUserID).Take(&sampleUser).Error; err != nil {
		t.Fatalf("load sample user: %v", err)
	}
	assertUserCount(t, db, "payment_resource", sampleUser.UserNo, 3)
	assertUserCount(t, db, "sub_category", sampleUser.UserNo, 7)
	assertUserCount(t, db, "hidden_sub_category", sampleUser.UserNo, 3)
	assertUserCount(t, db, "monthly_transaction", sampleUser.UserNo, 6)
	assertUserCount(t, db, "transaction", sampleUser.UserNo, 72)

	var firstTransaction transactionSchema
	if err := db.Where("user_no = ?", sampleUser.UserNo).Order("transaction_id").Take(&firstTransaction).Error; err != nil {
		t.Fatalf("load first sample transaction: %v", err)
	}
	mustExec(t, db, "UPDATE transaction SET transaction_name = ? WHERE transaction_id = ?", "手動変更", firstTransaction.TransactionID)
	if err := Run(ctx, db, databaseName, options); err != nil {
		t.Fatalf("repeat sample seed: %v", err)
	}
	assertUserCount(t, db, "transaction", sampleUser.UserNo, 72)
	var preservedName string
	if err := db.Table("transaction").Select("transaction_name").Where("transaction_id = ?", firstTransaction.TransactionID).Scan(&preservedName).Error; err != nil {
		t.Fatalf("read manually changed sample transaction: %v", err)
	}
	if preservedName != "手動変更" {
		t.Fatalf("sample data was overwritten: transaction_name=%q", preservedName)
	}
}

func createLegacyAuthSchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	userNoType := "BIGSERIAL PRIMARY KEY"
	mustExec(t, db, fmt.Sprintf(
		"CREATE TABLE %s (%s VARCHAR(64) NOT NULL, %s %s, %s VARCHAR(128) UNIQUE, %s TEXT)",
		quoteIdentifier("users"),
		quoteIdentifier("user_id"),
		quoteIdentifier("user_no"),
		userNoType,
		quoteIdentifier("email"),
		quoteIdentifier("PASSWORD"),
	))
	mustExec(t, db, fmt.Sprintf(
		"CREATE TABLE %s (%s BIGINT NOT NULL, %s VARCHAR(64) NOT NULL)",
		quoteIdentifier("user_token"),
		quoteIdentifier("user_no"),
		quoteIdentifier("token"),
	))
	insertLegacyUser := fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)",
		quoteIdentifier("users"),
		quoteIdentifier("user_id"),
		quoteIdentifier("user_no"),
		quoteIdentifier("email"),
		quoteIdentifier("PASSWORD"),
	)
	mustExec(t, db, insertLegacyUser, "duplicate-legacy-id", 98, "first@example.com", "password")
	mustExec(t, db, insertLegacyUser, "duplicate-legacy-id", 99, "second@example.com", "password")
	mustExec(t, db, "INSERT INTO user_token (user_no, token) VALUES (?, ?)", 99, "legacy-token")
}

func assertUserCount(t *testing.T, db *gorm.DB, table string, userNo uint64, want int64) {
	t.Helper()
	var count int64
	if err := db.Table(table).Where("user_no = ?", userNo).Count(&count).Error; err != nil {
		t.Fatalf("count %s for user %d: %v", table, userNo, err)
	}
	if count != want {
		t.Fatalf("%s count for user %d = %d, want %d", table, userNo, count, want)
	}
}

func exerciseLegacySchema(t *testing.T, db *gorm.DB, databaseName string) {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve integration test path")
	}
	directory := filepath.Join("psql", "init")
	fileName := "init.sql"
	legacyPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", directory, fileName)
	legacySQL, err := os.ReadFile(legacyPath)
	if err != nil {
		t.Fatalf("read legacy schema %s: %v", legacyPath, err)
	}
	if err := db.Exec(string(legacySQL)).Error; err != nil {
		t.Fatalf("apply legacy schema %s: %v", legacyPath, err)
	}
	// PostgreSQL sequences are not advanced by the fixed IDs in init.sql. Run the
	// normal startup repair once before loading the compatibility sample fixture.
	if err := Run(context.Background(), db, databaseName, Options{}); err != nil {
		t.Fatalf("migrate legacy schema before loading compatibility data: %v", err)
	}
	legacyDataPath := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", directory, "moneyhooks_data.sql")
	legacyDataSQL, err := os.ReadFile(legacyDataPath)
	if err != nil {
		t.Fatalf("read legacy data %s: %v", legacyDataPath, err)
	}
	if err := db.Exec(string(legacyDataSQL)).Error; err != nil {
		t.Fatalf("apply legacy data %s: %v", legacyDataPath, err)
	}
	var legacySampleTransactionCount int64
	if err := db.Table("transaction").Where("user_no = ?", 2).Count(&legacySampleTransactionCount).Error; err != nil {
		t.Fatalf("count legacy sample transactions: %v", err)
	}
	if legacySampleTransactionCount == 0 {
		t.Fatal("legacy sample data contains no transactions")
	}

	mustExec(t, db, "INSERT INTO users (user_id, user_no) VALUES (?, ?)", "legacy-preserved-user", 99)
	options := Options{
		EnableSeedData:    true,
		SeedReferenceTime: time.Date(2026, time.August, 29, 12, 0, 0, 0, time.FixedZone("Asia/Tokyo", 9*60*60)),
	}
	if err := Run(context.Background(), db, databaseName, options); err != nil {
		t.Fatalf("migrate legacy schema: %v", err)
	}
	assertCount(t, db, "category", 29)
	assertCount(t, db, "sub_category", 98)
	assertCount(t, db, "payment_type", 3)
	assertUserCount(t, db, "payment_resource", 2, 3)
	assertUserCount(t, db, "monthly_transaction", 2, 6)
	assertUserCount(t, db, "transaction", 2, legacySampleTransactionCount)

	var preservedCount int64
	if err := db.Table("users").Where("user_no = ? AND user_id = ?", 99, "legacy-preserved-user").Count(&preservedCount).Error; err != nil {
		t.Fatalf("read preserved legacy user: %v", err)
	}
	if preservedCount != 1 {
		t.Fatalf("legacy user was not preserved")
	}
	mustExec(t, db, "INSERT INTO users (user_id) VALUES (?)", "legacy-sequence-user")
	var generatedUserNo uint64
	if err := db.Table("users").Select("user_no").Where("user_id = ?", "legacy-sequence-user").Scan(&generatedUserNo).Error; err != nil {
		t.Fatalf("read legacy generated user number: %v", err)
	}
	if generatedUserNo <= 99 {
		t.Fatalf("legacy generated user number = %d, want greater than 99", generatedUserNo)
	}

	if err := Run(context.Background(), db, databaseName, options); err != nil {
		t.Fatalf("migrate legacy database with existing sample user: %v", err)
	}
	assertUserCount(t, db, "payment_resource", 2, 3)
	assertUserCount(t, db, "monthly_transaction", 2, 6)
	assertUserCount(t, db, "transaction", 2, legacySampleTransactionCount)
}

func assertCount(t *testing.T, db *gorm.DB, table string, want int64) {
	t.Helper()
	var count int64
	if err := db.Table(table).Count(&count).Error; err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if count != want {
		t.Fatalf("%s count = %d, want %d", table, count, want)
	}
}

func mustExec(t *testing.T, db *gorm.DB, query string, args ...any) {
	t.Helper()
	if err := db.Exec(query, args...).Error; err != nil {
		t.Fatalf("execute %q: %v", query, err)
	}
}

func closeGormDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
