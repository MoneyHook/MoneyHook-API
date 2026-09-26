//go:build integration

package store_postgres

import (
	"MoneyHook/MoneyHook-API/model"
	subcategorydomain "MoneyHook/MoneyHook-API/subcategory"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Tests use an isolated schema, never application tables.
func legacyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("MIGRATION_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("MIGRATION_TEST_POSTGRES_DSN is not set")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("legacy_write_test_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Exec("DROP SCHEMA " + schema + " CASCADE"); sqlDB, _ := admin.DB(); sqlDB.Close() })
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB, _ := db.DB(); sqlDB.Close() })
	statements := []string{
		`CREATE TABLE category (category_id BIGINT PRIMARY KEY, category_name VARCHAR(16) NOT NULL, order_num INT NOT NULL DEFAULT 0)`,
		`CREATE TABLE sub_category (sub_category_id BIGSERIAL PRIMARY KEY, user_no BIGINT NOT NULL, category_id BIGINT NOT NULL, sub_category_name VARCHAR(16) NOT NULL, UNIQUE(user_no,category_id,sub_category_name))`,
		`CREATE TABLE hidden_sub_category (user_no BIGINT NOT NULL, sub_category_id BIGINT NOT NULL, UNIQUE(user_no,sub_category_id))`,
		`CREATE TABLE payment_resource (payment_id BIGSERIAL PRIMARY KEY, user_no BIGINT NOT NULL, payment_name VARCHAR(32) NOT NULL)`,
		`CREATE TABLE "transaction" (transaction_id BIGSERIAL PRIMARY KEY, user_no BIGINT NOT NULL, transaction_name VARCHAR(32) NOT NULL, transaction_amount BIGINT NOT NULL, transaction_date DATE NOT NULL, transaction_time TIME, category_id BIGINT NOT NULL, sub_category_id BIGINT NOT NULL REFERENCES sub_category(sub_category_id), fixed_flg BOOLEAN NOT NULL, payment_id BIGINT)`,
		`CREATE TABLE monthly_transaction (monthly_transaction_id BIGSERIAL PRIMARY KEY, user_no BIGINT NOT NULL, monthly_transaction_name VARCHAR(32) NOT NULL, monthly_transaction_amount BIGINT NOT NULL, monthly_transaction_date INTEGER NOT NULL CHECK(monthly_transaction_date BETWEEN 1 AND 31), category_id BIGINT NOT NULL, sub_category_id BIGINT NOT NULL REFERENCES sub_category(sub_category_id), include_flg BOOLEAN NOT NULL, payment_id BIGINT)`,
		`INSERT INTO category (category_id, category_name) VALUES (1, '食費'), (2, '日用品')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func v1Row() model.V1TransactionWrite {
	return model.V1TransactionWrite{
		UserId:          "2",
		TransactionDate: "2026-09-25",
		TransactionName: "ランチ",
		Amount:          1200,
		Sign:            -1,
		CategoryId:      "1",
		SubCategoryName: "外食",
	}
}

func TestV1CreateTransactionResolvesAndExposesSubcategories(t *testing.T) {
	db := legacyTestDB(t)
	store := NewTransactionStore(db)
	if err := db.Exec(`INSERT INTO sub_category (sub_category_id, user_no, category_id, sub_category_name) VALUES (10, 1, 1, '外食')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO hidden_sub_category (user_no, sub_category_id) VALUES (2, 10)`).Error; err != nil {
		t.Fatal(err)
	}

	created, err := store.CreateV1Transaction(ptr(v1Row()))
	if err != nil {
		t.Fatal(err)
	}
	if created.SubCategoryId != "10" || created.SubCategoryName != "外食" {
		t.Fatalf("unexpected subcategory: %#v", created)
	}
	assertLegacyCount(t, db, "sub_category", 1)
	assertLegacyCount(t, db, "hidden_sub_category", 0)
}

func TestV1CreateTransactionPrefersEnabledThenMasterSubcategory(t *testing.T) {
	db := legacyTestDB(t)
	store := NewTransactionStore(db)
	if err := db.Exec(`INSERT INTO sub_category (sub_category_id, user_no, category_id, sub_category_name) VALUES (10, 1, 1, '外食'), (11, 2, 1, '外食')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO hidden_sub_category (user_no, sub_category_id) VALUES (2, 10)`).Error; err != nil {
		t.Fatal(err)
	}

	created, err := store.CreateV1Transaction(ptr(v1Row()))
	if err != nil {
		t.Fatal(err)
	}
	if created.SubCategoryId != "11" {
		t.Fatalf("enabled user subcategory should win, got %s", created.SubCategoryId)
	}
}

func TestV1CreateTransactionRollsBackNewSubcategory(t *testing.T) {
	db := legacyTestDB(t)
	store := NewTransactionStore(db)
	input := v1Row()
	input.TransactionDate = "invalid"
	if _, err := store.CreateV1Transaction(&input); err == nil {
		t.Fatal("invalid transaction should fail")
	}
	assertLegacyCount(t, db, "sub_category", 0)
	assertLegacyCount(t, db, "transaction", 0)
}

func TestV1CreateTransactionRejectsAnotherUsersSubcategory(t *testing.T) {
	db := legacyTestDB(t)
	store := NewTransactionStore(db)
	if err := db.Exec(`INSERT INTO sub_category (sub_category_id, user_no, category_id, sub_category_name) VALUES (12, 3, 1, '他人')`).Error; err != nil {
		t.Fatal(err)
	}
	input := v1Row()
	input.SubCategoryName = ""
	input.SubCategoryId = "12"
	if _, err := store.CreateV1Transaction(&input); !errors.Is(err, transactiondomain.ErrInvalidRelation) {
		t.Fatalf("expected invalid relation, got %v", err)
	}
}

func TestSubCategoryVisibilityIsUserScopedAndIdempotent(t *testing.T) {
	db := legacyTestDB(t)
	if err := db.Exec(`INSERT INTO sub_category (sub_category_id, user_no, category_id, sub_category_name) VALUES (10, 1, 1, '外食'), (11, 3, 1, '他人')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO hidden_sub_category (user_no, sub_category_id) VALUES (3, 10)`).Error; err != nil {
		t.Fatal(err)
	}

	categoryStore := NewCategoryStore(db)
	categories := categoryStore.GetCategoryWithSubCategoryList("2")
	if len(*categories) == 0 || len((*categories)[0].SubCategoryList) == 0 || !(*categories)[0].SubCategoryList[0].Enable {
		t.Fatalf("another user's hidden marker must not affect user 2: %#v", categories)
	}

	store := NewSubCategoryStore(db)
	visibility := model.EditSubCategoryModel{UserId: "2", SubCategoryId: "10", IsEnable: false}
	if err := store.HideSubCategory(&visibility); err != nil {
		t.Fatal(err)
	}
	if err := store.HideSubCategory(&visibility); err != nil {
		t.Fatalf("repeated hide must be idempotent: %v", err)
	}
	var count int64
	db.Table("hidden_sub_category").Where("user_no = ?", "2").Count(&count)
	if count != 1 {
		t.Fatalf("hidden marker count=%d want=1", count)
	}
	if err := store.ExposeSubCategory(&visibility); err != nil {
		t.Fatal(err)
	}
	if err := store.ExposeSubCategory(&visibility); err != nil {
		t.Fatalf("repeated expose must be idempotent: %v", err)
	}

	visibility.SubCategoryId = "11"
	if err := store.HideSubCategory(&visibility); !errors.Is(err, subcategorydomain.ErrNotFound) {
		t.Fatalf("another user's subcategory must be rejected, got %v", err)
	}
}

func ptr[T any](value T) *T { return &value }
func legacyRow(name string) model.AddTransaction {
	return model.AddTransaction{UserId: "2", TransactionDate: "2026-09-01", TransactionAmount: -1200, TransactionName: "ランチ", CategoryId: "2", SubCategoryName: name}
}
func assertLegacyCount(t *testing.T, db *gorm.DB, table string, want int64) {
	t.Helper()
	var count int64
	if err := db.Table(table).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("%s count=%d want=%d", table, count, want)
	}
}
func TestLegacyWritesRollbackSubcategories(t *testing.T) {
	db := legacyTestDB(t)
	store := NewTransactionStore(db)
	invalid := legacyRow("失敗")
	invalid.TransactionDate = "invalid-date"
	if err := store.AddTransaction(&invalid); err == nil {
		t.Fatal("single insert should fail")
	}
	assertLegacyCount(t, db, "sub_category", 0)
	assertLegacyCount(t, db, "transaction", 0)
	if invalid.SubCategoryId != "" {
		t.Fatal("failed insert mutated caller's subcategory ID")
	}
	batch := model.AddTransactionList{UserId: "2", TransactionList: []model.AddTransaction{legacyRow("正常行"), invalid}}
	if err := store.AddTransactionList(&batch); err == nil {
		t.Fatal("batch should fail")
	}
	assertLegacyCount(t, db, "sub_category", 0)
	assertLegacyCount(t, db, "transaction", 0)
	// Existing subcategories survive a failed edit; newly created ones do not.
	valid := legacyRow("既存")
	if err := store.AddTransaction(&valid); err != nil {
		t.Fatal(err)
	}
	var id string
	if err := db.Table("transaction").Select("transaction_id").Take(&id).Error; err != nil {
		t.Fatal(err)
	}
	edit := model.EditTransaction{UserId: "2", TransactionId: id, TransactionDate: "invalid-date", TransactionAmount: -1, TransactionName: "変更", CategoryId: "2", SubCategoryName: "新規"}
	if err := store.EditTransaction(&edit); err == nil {
		t.Fatal("edit should fail")
	}
	assertLegacyCount(t, db, "sub_category", 1)
	assertLegacyCount(t, db, "transaction", 1)
	var name string
	db.Table("transaction").Select("transaction_name").Take(&name)
	if name != "ランチ" {
		t.Fatalf("failed edit changed name: %s", name)
	}
}
func TestLegacyWritesReuseSubcategoriesAndPropagateFailure(t *testing.T) {
	db := legacyTestDB(t)
	store := NewTransactionStore(db)
	input := model.AddTransactionList{UserId: "2", TransactionList: []model.AddTransaction{legacyRow("共通"), legacyRow("共通")}}
	if err := store.AddTransactionList(&input); err != nil {
		t.Fatal(err)
	}
	assertLegacyCount(t, db, "sub_category", 1)
	assertLegacyCount(t, db, "transaction", 2)
	var id string
	db.Table("transaction").Select("transaction_id").Take(&id)
	edit := model.EditTransaction{UserId: "2", TransactionId: id, TransactionDate: "2026-09-02", TransactionAmount: -5, TransactionName: "変更", CategoryId: "2", SubCategoryName: "共通"}
	if err := store.EditTransaction(&edit); err != nil {
		t.Fatal(err)
	}
	assertLegacyCount(t, db, "sub_category", 1)
	edit.SubCategoryName = strings.Repeat("あ", 17)
	if err := store.EditTransaction(&edit); !errors.Is(err, subcategorydomain.ErrResolveFailed) {
		t.Fatalf("expected subcategory failure, got %v", err)
	}
	assertLegacyCount(t, db, "sub_category", 1)
}
func TestLegacyConcurrentSubcategoryCreation(t *testing.T) {
	db := legacyTestDB(t)
	store := NewTransactionStore(db)
	var wg sync.WaitGroup
	failures := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); row := legacyRow("同時作成"); failures <- store.AddTransaction(&row) }()
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	assertLegacyCount(t, db, "sub_category", 1)
	assertLegacyCount(t, db, "transaction", 2)
}
func TestLegacyFixedWritesResolveAndRollbackSubcategories(t *testing.T) {
	db := legacyTestDB(t)
	store := NewFixedStore(db)
	input := model.AddFixed{UserId: "2", CategoryId: "2", SubCategoryName: "固定", MonthlyTransactionName: "家賃", MonthlyTransactionAmount: -50000, MonthlyTransactionDate: 1}
	if err := store.AddFixed(&input); err != nil {
		t.Fatal(err)
	}
	if err := store.AddFixed(&input); err != nil {
		t.Fatal(err)
	}
	assertLegacyCount(t, db, "sub_category", 1)
	assertLegacyCount(t, db, "monthly_transaction", 2)
	input.SubCategoryName = "失敗"
	input.MonthlyTransactionDate = 99
	if err := store.AddFixed(&input); err == nil {
		t.Fatal("invalid day should fail")
	}
	assertLegacyCount(t, db, "sub_category", 1)
	var id string
	db.Table("monthly_transaction").Select("monthly_transaction_id").Take(&id)
	edit := model.EditFixed{UserId: "2", MonthlyTransactionId: id, CategoryId: "2", SubCategoryName: "固定", MonthlyTransactionName: "家賃変更", MonthlyTransactionAmount: -60000, MonthlyTransactionDate: 2, IncludeFlg: true}
	if err := store.EditFixed(&edit); err != nil {
		t.Fatal(err)
	}
	assertLegacyCount(t, db, "sub_category", 1)
	edit.SubCategoryName = "新規"
	edit.MonthlyTransactionDate = 99
	if err := store.EditFixed(&edit); err == nil {
		t.Fatal("invalid edit should fail")
	}
	assertLegacyCount(t, db, "sub_category", 1)
	edit.SubCategoryName = strings.Repeat("あ", 17)
	edit.MonthlyTransactionDate = 2
	if err := store.EditFixed(&edit); !errors.Is(err, subcategorydomain.ErrResolveFailed) {
		t.Fatalf("expected subcategory failure, got %v", err)
	}
}

func TestLegacyMissingEditsDoNotLeaveSubcategories(t *testing.T) {
	db := legacyTestDB(t)
	err := NewTransactionStore(db).EditTransaction(&model.EditTransaction{UserId: "2", TransactionId: "999", CategoryId: "2", SubCategoryName: "不存在", TransactionName: "取引", TransactionDate: "2026-09-01"})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing transaction: %v", err)
	}
	assertLegacyCount(t, db, "sub_category", 0)
	err = NewFixedStore(db).EditFixed(&model.EditFixed{UserId: "2", MonthlyTransactionId: "999", CategoryId: "2", SubCategoryName: "不存在", MonthlyTransactionName: "固定費", MonthlyTransactionDate: 1})
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("missing fixed transaction: %v", err)
	}
	assertLegacyCount(t, db, "sub_category", 0)
}
