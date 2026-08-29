package migration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type schemaChange struct {
	Action string
	Table  string
	Column string
}

func detectMissingSchema(ctx context.Context, db *gorm.DB, models []any) ([]schemaChange, error) {
	expectedColumns, err := expectedSchemaColumns(db, models)
	if err != nil {
		return nil, err
	}

	var changes []schemaChange
	for table, expected := range expectedColumns {
		if !db.WithContext(ctx).Migrator().HasTable(table) {
			changes = append(changes, schemaChange{Action: "create_table", Table: table})
			continue
		}
		actual, err := loadColumns(ctx, db, table)
		if err != nil {
			return nil, err
		}
		for column := range expected {
			if _, exists := actual[column]; !exists {
				changes = append(changes, schemaChange{Action: "add_column", Table: table, Column: column})
			}
		}
	}
	return changes, nil
}

func synchronizeSequences(ctx context.Context, db *gorm.DB, dialect Dialect) error {
	for _, requirement := range sequenceRequirements {
		var maximum sql.NullInt64
		maxQuery := fmt.Sprintf(
			"SELECT MAX(%s) FROM %s",
			quoteIdentifier(dialect, requirement.Column),
			quoteIdentifier(dialect, requirement.Table),
		)
		if err := db.WithContext(ctx).Raw(maxQuery).Scan(&maximum).Error; err != nil {
			return err
		}

		switch dialect {
		case PostgreSQL:
			var sequenceName sql.NullString
			if err := db.WithContext(ctx).
				Raw("SELECT pg_get_serial_sequence(?, ?)", requirement.Table, requirement.Column).
				Scan(&sequenceName).Error; err != nil {
				return err
			}
			if !sequenceName.Valid || sequenceName.String == "" {
				return fmt.Errorf("serial sequence not found for %s.%s", requirement.Table, requirement.Column)
			}
			value := int64(1)
			isCalled := false
			if maximum.Valid {
				value = maximum.Int64
				isCalled = true
			}
			if err := db.WithContext(ctx).
				Exec("SELECT setval(CAST(? AS regclass), ?, ?)", sequenceName.String, value, isCalled).Error; err != nil {
				return err
			}
		case MySQL:
			nextValue := int64(1)
			if maximum.Valid {
				nextValue = maximum.Int64 + 1
			}
			query := fmt.Sprintf("ALTER TABLE %s AUTO_INCREMENT = %d", quoteIdentifier(dialect, requirement.Table), nextValue)
			if err := db.WithContext(ctx).Exec(query).Error; err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported database dialect %q", dialect)
		}
	}
	return nil
}

func validateSchema(ctx context.Context, db *gorm.DB, dialect Dialect, models []any) error {
	expectedColumns, err := expectedSchemaColumns(db, models)
	if err != nil {
		return err
	}

	for table, expected := range expectedColumns {
		if !db.WithContext(ctx).Migrator().HasTable(table) {
			return fmt.Errorf("required table %s is missing", table)
		}
		actual, err := loadColumns(ctx, db, table)
		if err != nil {
			return err
		}
		for column := range expected {
			if _, exists := actual[column]; !exists {
				return fmt.Errorf("required column %s.%s is missing", table, column)
			}
		}
	}

	for _, requirement := range uniqueRequirements {
		indexes, err := loadUniqueIndexes(ctx, db, dialect, requirement.Table)
		if err != nil {
			return err
		}
		if !containsUniqueIndex(indexes, requirement.Columns) {
			return fmt.Errorf("required unique index on %s(%s) is missing", requirement.Table, strings.Join(requirement.Columns, ","))
		}
	}
	for _, requirement := range foreignKeyRequirements {
		keys, err := loadForeignKeys(ctx, db, dialect, requirement.Table)
		if err != nil {
			return err
		}
		if !containsForeignKey(keys, requirement) {
			return fmt.Errorf("required foreign key %s.%s -> %s.%s is missing", requirement.Table, requirement.Column, requirement.ReferencedTable, requirement.ReferencedColumn)
		}
	}
	return nil
}

func logPreservedDrift(ctx context.Context, db *gorm.DB, dialect Dialect, models []any) error {
	expectedColumns, err := expectedSchemaColumns(db, models)
	if err != nil {
		return err
	}

	tables, err := db.WithContext(ctx).Migrator().GetTables()
	if err != nil {
		return err
	}
	sort.Strings(tables)
	for _, table := range tables {
		expected, managed := expectedColumns[table]
		if !managed {
			log.Printf("event=schema_migration_warning type=extra_table table=%s action=preserved", table)
			continue
		}
		actual, err := loadColumns(ctx, db, table)
		if err != nil {
			return err
		}
		for column := range actual {
			if _, exists := expected[column]; !exists {
				log.Printf("event=schema_migration_warning type=extra_column table=%s column=%s action=preserved", table, column)
			}
		}

		indexes, err := loadUniqueIndexes(ctx, db, dialect, table)
		if err != nil {
			return err
		}
		for _, index := range indexes {
			if !isExpectedUniqueIndex(table, index.Columns) {
				log.Printf("event=schema_migration_warning type=extra_unique_index table=%s index=%s columns=%s action=preserved", table, index.Name, strings.Join(index.Columns, ","))
			}
		}

		keys, err := loadForeignKeys(ctx, db, dialect, table)
		if err != nil {
			return err
		}
		for _, key := range keys {
			if !isExpectedForeignKey(table, key) {
				log.Printf("event=schema_migration_warning type=extra_foreign_key table=%s constraint=%s column=%s action=preserved", table, key.Name, key.Column)
			}
		}
	}
	return nil
}

func expectedSchemaColumns(db *gorm.DB, models []any) (map[string]map[string]struct{}, error) {
	result := make(map[string]map[string]struct{}, len(models))
	for _, model := range models {
		statement := &gorm.Statement{DB: db}
		if err := statement.Parse(model); err != nil {
			return nil, err
		}
		columns := make(map[string]struct{}, len(statement.Schema.DBNames))
		for _, column := range statement.Schema.DBNames {
			columns[column] = struct{}{}
		}
		result[statement.Schema.Table] = columns
	}
	return result, nil
}

func loadColumns(ctx context.Context, db *gorm.DB, table string) (map[string]struct{}, error) {
	columnTypes, err := db.WithContext(ctx).Migrator().ColumnTypes(table)
	if err != nil {
		return nil, err
	}
	columns := make(map[string]struct{}, len(columnTypes))
	for _, columnType := range columnTypes {
		columns[columnType.Name()] = struct{}{}
	}
	return columns, nil
}

func isExpectedUniqueIndex(table string, columns []string) bool {
	for _, requirement := range uniqueRequirements {
		if requirement.Table == table && sameStringSet(requirement.Columns, columns) {
			return true
		}
	}
	return false
}

func isExpectedForeignKey(table string, key foreignKey) bool {
	for _, requirement := range foreignKeyRequirements {
		if requirement.Table == table && containsForeignKey([]foreignKey{key}, requirement) {
			return true
		}
	}
	return false
}
