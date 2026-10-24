//go:build integration

package store_postgres

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"testing"
)

func TestCategoryWithSubcategoriesUsesTwoQueriesIntegration(t *testing.T) {
	db := legacyTestDB(t)
	for _, stmt := range []string{
		`INSERT INTO category (category_id,category_name,order_num) VALUES (3,'空のカテゴリ',-1)`,
		`INSERT INTO sub_category (sub_category_id,user_no,category_id,sub_category_name) VALUES
 (1,1,1,'共通'),(2,2,1,'自分'),(3,3,1,'他人'),(4,1,2,'非表示')`,
		`INSERT INTO hidden_sub_category (user_no,sub_category_id) VALUES (2,4),(3,1)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatal(err)
		}
	}
	c := &householdReadCounter{Interface: logger.Default}
	result := NewCategoryStore(db.Session(&gorm.Session{Logger: c})).GetCategoryWithSubCategoryList("2")
	if c.count != 2 || len(*result) != 3 {
		t.Fatalf("queries=%d categories=%d", c.count, len(*result))
	}
	if (*result)[0].CategoryId != "3" || (*result)[0].SubCategoryList == nil || len((*result)[0].SubCategoryList) != 0 {
		t.Fatal("empty category or order lost")
	}
	seen := map[string]bool{}
	for _, category := range *result {
		for _, sub := range category.SubCategoryList {
			seen[sub.SubCategoryId] = sub.Enable
		}
	}
	if len(seen) != 3 || !seen["1"] || !seen["2"] || seen["4"] {
		t.Fatal(seen)
	}
}
