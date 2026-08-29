package migration

import (
	"fmt"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestMasterDataDefinitions(t *testing.T) {
	if got, want := len(masterCategories), 29; got != want {
		t.Fatalf("master category count = %d, want %d", got, want)
	}
	if got, want := len(masterSubCategories), 91; got != want {
		t.Fatalf("master sub-category count = %d, want %d", got, want)
	}
	if got, want := len(masterPaymentTypes), 3; got != want {
		t.Fatalf("master payment type count = %d, want %d", got, want)
	}

	categoryIDs := make(map[uint64]struct{}, len(masterCategories))
	for _, category := range masterCategories {
		if _, duplicated := categoryIDs[category.CategoryID]; duplicated {
			t.Fatalf("duplicated category id: %d", category.CategoryID)
		}
		categoryIDs[category.CategoryID] = struct{}{}
	}

	subCategoryKeys := make(map[string]struct{}, len(masterSubCategories))
	hasDefault := make(map[uint64]bool, len(masterCategories))
	for _, subCategory := range masterSubCategories {
		if _, exists := categoryIDs[subCategory.CategoryID]; !exists {
			t.Fatalf("sub-category references undefined category: %d", subCategory.CategoryID)
		}
		key := fmt.Sprintf("%d:%s", subCategory.CategoryID, subCategory.Name)
		if _, duplicated := subCategoryKeys[key]; duplicated {
			t.Fatalf("duplicated sub-category: %s", key)
		}
		subCategoryKeys[key] = struct{}{}
		if subCategory.Name == "なし" {
			hasDefault[subCategory.CategoryID] = true
		}
	}
	for categoryID := range categoryIDs {
		if !hasDefault[categoryID] {
			t.Errorf("category %d has no default sub-category", categoryID)
		}
	}
}

func TestSchemaModelCoverage(t *testing.T) {
	wantTables := map[string]struct{}{
		"users": {}, "category": {}, "sub_category": {},
		"hidden_sub_category": {}, "payment_type": {}, "payment_resource": {},
		"transaction": {}, "monthly_transaction": {},
	}

	gotTables := make(map[string]struct{}, len(schemaModels))
	for _, model := range schemaModels {
		parsed, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			t.Fatalf("parse schema model: %v", err)
		}
		if _, duplicated := gotTables[parsed.Table]; duplicated {
			t.Fatalf("duplicated schema table: %s", parsed.Table)
		}
		gotTables[parsed.Table] = struct{}{}
	}

	if len(gotTables) != len(wantTables) {
		t.Fatalf("schema model table count = %d, want %d", len(gotTables), len(wantTables))
	}
	for table := range wantTables {
		if _, exists := gotTables[table]; !exists {
			t.Errorf("missing schema model for table %s", table)
		}
	}
}

func TestSameStringSet(t *testing.T) {
	if !sameStringSet([]string{"user_no", "category_id"}, []string{"category_id", "user_no"}) {
		t.Fatal("sameStringSet should ignore column order")
	}
	if sameStringSet([]string{"user_no"}, []string{"category_id"}) {
		t.Fatal("sameStringSet accepted different columns")
	}
}
