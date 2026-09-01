package migration

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

var budgetPrimaryKeyColumns = []string{"user_no", "effective_from"}

// migrateLegacyBudgetSchema converts the first version of the budget table,
// which stored one row per user, into the effective-date history shape. The
// old start_month column is preserved as nullable drift so this migration
// never drops user data while new writes can omit the obsolete column.
func migrateLegacyBudgetSchema(ctx context.Context, db *gorm.DB) error {
	if !db.WithContext(ctx).Migrator().HasTable("budget") {
		return nil
	}

	columns, err := loadColumns(ctx, db, "budget")
	if err != nil {
		return err
	}
	_, hasEffectiveFrom := columns["effective_from"]
	_, hasLegacyStartMonth := columns["start_month"]
	if !hasEffectiveFrom {
		if !hasLegacyStartMonth {
			return fmt.Errorf("budget table has neither effective_from nor legacy start_month")
		}

		addColumn := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s DATE", quoteIdentifier("budget"), quoteIdentifier("effective_from"))
		if err := db.WithContext(ctx).Exec(addColumn).Error; err != nil {
			return fmt.Errorf("add budget.effective_from: %w", err)
		}
		hasEffectiveFrom = true
		log.Printf("event=schema_migration_change action=add_column table=budget column=effective_from")
	}
	if hasEffectiveFrom && hasLegacyStartMonth {
		copyColumn := fmt.Sprintf("UPDATE %s SET %s = %s WHERE %s IS NULL", quoteIdentifier("budget"), quoteIdentifier("effective_from"), quoteIdentifier("start_month"), quoteIdentifier("effective_from"))
		if err := db.WithContext(ctx).Exec(copyColumn).Error; err != nil {
			return fmt.Errorf("copy budget.start_month to budget.effective_from: %w", err)
		}
	}
	if hasLegacyStartMonth {
		dropNotNull := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL", quoteIdentifier("budget"), quoteIdentifier("start_month"))
		if err := db.WithContext(ctx).Exec(dropNotNull).Error; err != nil {
			return fmt.Errorf("make legacy budget.start_month nullable: %w", err)
		}
		log.Printf("event=schema_migration_change action=relax_legacy_column table=budget column=start_month")
	}

	setNotNull := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL", quoteIdentifier("budget"), quoteIdentifier("effective_from"))
	if err := db.WithContext(ctx).Exec(setNotNull).Error; err != nil {
		return fmt.Errorf("make budget.effective_from non-null: %w", err)
	}

	primaryKey, err := loadPrimaryKeyColumns(ctx, db, "budget")
	if err != nil {
		return err
	}
	if sameStringSlice(primaryKey, budgetPrimaryKeyColumns) {
		return nil
	}

	constraintName, err := loadPrimaryKeyConstraintName(ctx, db, "budget")
	if err != nil {
		return err
	}
	if constraintName != "" {
		dropConstraint := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", quoteIdentifier("budget"), quoteIdentifier(constraintName))
		if err := db.WithContext(ctx).Exec(dropConstraint).Error; err != nil {
			return fmt.Errorf("drop legacy budget primary key: %w", err)
		}
	}
	addConstraint := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)", quoteIdentifier("budget"), quoteIdentifier("pk_budget_user_effective_from"), quoteIdentifiers(budgetPrimaryKeyColumns))
	if err := db.WithContext(ctx).Exec(addConstraint).Error; err != nil {
		return fmt.Errorf("create budget effective-date primary key: %w", err)
	}
	log.Printf("event=schema_migration_change action=replace_primary_key table=budget columns=%s", strings.Join(budgetPrimaryKeyColumns, ","))
	return nil
}

func loadPrimaryKeyColumns(ctx context.Context, db *gorm.DB, table string) ([]string, error) {
	query := `
		SELECT attribute.attname
		FROM pg_index index_definition
		JOIN pg_class relation ON relation.oid = index_definition.indrelid
		JOIN pg_namespace namespace ON namespace.oid = relation.relnamespace
		JOIN pg_attribute attribute ON attribute.attrelid = relation.oid AND attribute.attnum = ANY(index_definition.indkey)
		WHERE namespace.nspname = current_schema() AND relation.relname = ? AND index_definition.indisprimary
		ORDER BY array_position(index_definition.indkey, attribute.attnum)`

	rows, err := db.WithContext(ctx).Raw(query, table).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

func loadPrimaryKeyConstraintName(ctx context.Context, db *gorm.DB, table string) (string, error) {
	var name string
	err := db.WithContext(ctx).Raw(`
		SELECT constraint_definition.conname
		FROM pg_constraint constraint_definition
		JOIN pg_class relation ON relation.oid = constraint_definition.conrelid
		JOIN pg_namespace namespace ON namespace.oid = relation.relnamespace
		WHERE namespace.nspname = current_schema() AND relation.relname = ? AND constraint_definition.contype = 'p'
		LIMIT 1`, table).Scan(&name).Error
	return name, err
}

func sameStringSlice(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
