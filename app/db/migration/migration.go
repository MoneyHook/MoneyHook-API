package migration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Dialect string

const (
	MySQL      Dialect = "mysql"
	PostgreSQL Dialect = "postgresql"
)

const migrationTimeout = 60 * time.Second

type Options struct {
	EnableSeedData    bool
	SeedReferenceTime time.Time
}

func Run(parent context.Context, db *gorm.DB, dialect Dialect, databaseName string, options Options) error {
	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(parent, migrationTimeout)
	defer cancel()

	log.Printf("event=schema_migration_start dialect=%s database=%s", dialect, databaseName)

	release, err := acquireLock(ctx, db, dialect, databaseName)
	if err != nil {
		return fmt.Errorf("acquire schema migration lock: %w", err)
	}
	defer release()

	migrationDB := db.WithContext(ctx)
	if err := removeLegacyAuthSchema(ctx, migrationDB, dialect); err != nil {
		return fmt.Errorf("remove legacy authentication schema: %w", err)
	}
	models, err := schemaModelsForDialect(ctx, migrationDB, dialect)
	if err != nil {
		return fmt.Errorf("select schema models: %w", err)
	}
	pendingChanges, err := detectMissingSchema(ctx, migrationDB, models)
	if err != nil {
		return fmt.Errorf("inspect schema before migration: %w", err)
	}
	if err := migrationDB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate schema: %w", err)
	}
	for _, change := range pendingChanges {
		log.Printf("event=schema_migration_change action=%s table=%s column=%s", change.Action, change.Table, change.Column)
	}

	if err := ensureUniqueIndexes(ctx, migrationDB, dialect); err != nil {
		return fmt.Errorf("reconcile unique indexes: %w", err)
	}
	if err := ensureForeignKeys(ctx, migrationDB, dialect); err != nil {
		return fmt.Errorf("reconcile foreign keys: %w", err)
	}
	if err := seedMasterData(ctx, migrationDB); err != nil {
		return fmt.Errorf("seed master data: %w", err)
	}
	if err := synchronizeSequences(ctx, migrationDB, dialect); err != nil {
		return fmt.Errorf("synchronize auto increment values after master seed: %w", err)
	}
	if err := seedSampleData(ctx, migrationDB, options); err != nil {
		return fmt.Errorf("seed sample data: %w", err)
	}
	if err := synchronizeSequences(ctx, migrationDB, dialect); err != nil {
		return fmt.Errorf("synchronize auto increment values: %w", err)
	}
	if err := validateSchema(ctx, migrationDB, dialect, models); err != nil {
		return fmt.Errorf("validate migrated schema: %w", err)
	}
	if err := logPreservedDrift(ctx, migrationDB, dialect, models); err != nil {
		log.Printf("event=schema_migration_warning type=drift_inspection_failed error=%q", err)
	}

	log.Printf("event=schema_migration_complete dialect=%s database=%s duration_ms=%d", dialect, databaseName, time.Since(startedAt).Milliseconds())
	return nil
}

func acquireLock(ctx context.Context, db *gorm.DB, dialect Dialect, databaseName string) (func(), error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return nil, err
	}

	lockDigest := sha256.Sum256([]byte(databaseName))
	lockName := "moneyhooks:schema:" + hex.EncodeToString(lockDigest[:16])
	locked := false

	switch dialect {
	case PostgreSQL:
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for !locked {
			if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(hashtext($1))", lockName).Scan(&locked); err != nil {
				conn.Close()
				return nil, err
			}
			if locked {
				break
			}
			select {
			case <-ctx.Done():
				conn.Close()
				return nil, ctx.Err()
			case <-ticker.C:
			}
		}
	case MySQL:
		var result sql.NullInt64
		if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", lockName, int(migrationTimeout.Seconds())).Scan(&result); err != nil {
			conn.Close()
			return nil, err
		}
		locked = result.Valid && result.Int64 == 1
		if !locked {
			conn.Close()
			return nil, errors.New("timed out waiting for MySQL named lock")
		}
	default:
		conn.Close()
		return nil, fmt.Errorf("unsupported database dialect %q", dialect)
	}

	log.Printf("event=schema_migration_lock_acquired dialect=%s database=%s", dialect, databaseName)
	return func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		switch dialect {
		case PostgreSQL:
			var released bool
			if err := conn.QueryRowContext(releaseCtx, "SELECT pg_advisory_unlock(hashtext($1))", lockName).Scan(&released); err != nil || !released {
				log.Printf("event=schema_migration_warning type=lock_release_failed dialect=%s error=%q", dialect, err)
			}
		case MySQL:
			var released sql.NullInt64
			if err := conn.QueryRowContext(releaseCtx, "SELECT RELEASE_LOCK(?)", lockName).Scan(&released); err != nil || !released.Valid || released.Int64 != 1 {
				log.Printf("event=schema_migration_warning type=lock_release_failed dialect=%s error=%q", dialect, err)
			}
		}
		if err := conn.Close(); err != nil {
			log.Printf("event=schema_migration_warning type=lock_connection_close_failed dialect=%s error=%q", dialect, err)
		}
	}, nil
}

type uniqueIndex struct {
	Name    string
	Columns []string
}

type foreignKey struct {
	Name             string
	Column           string
	ReferencedTable  string
	ReferencedColumn string
}

func ensureUniqueIndexes(ctx context.Context, db *gorm.DB, dialect Dialect) error {
	for _, requirement := range uniqueRequirements {
		indexes, err := loadUniqueIndexes(ctx, db, dialect, requirement.Table)
		if err != nil {
			return err
		}
		if containsUniqueIndex(indexes, requirement.Columns) {
			continue
		}
		query := fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s)",
			quoteIdentifier(dialect, requirement.Name),
			quoteIdentifier(dialect, requirement.Table),
			quoteIdentifiers(dialect, requirement.Columns),
		)
		if err := db.WithContext(ctx).Exec(query).Error; err != nil {
			return fmt.Errorf("create unique index %s: %w", requirement.Name, err)
		}
		log.Printf("event=schema_migration_change action=create_unique_index table=%s columns=%s", requirement.Table, strings.Join(requirement.Columns, ","))
	}
	return nil
}

func ensureForeignKeys(ctx context.Context, db *gorm.DB, dialect Dialect) error {
	for _, requirement := range foreignKeyRequirements {
		keys, err := loadForeignKeys(ctx, db, dialect, requirement.Table)
		if err != nil {
			return err
		}
		if containsForeignKey(keys, requirement) {
			continue
		}
		query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
			quoteIdentifier(dialect, requirement.Table),
			quoteIdentifier(dialect, requirement.Name),
			quoteIdentifier(dialect, requirement.Column),
			quoteIdentifier(dialect, requirement.ReferencedTable),
			quoteIdentifier(dialect, requirement.ReferencedColumn),
		)
		if err := db.WithContext(ctx).Exec(query).Error; err != nil {
			return fmt.Errorf("create foreign key %s: %w", requirement.Name, err)
		}
		log.Printf("event=schema_migration_change action=create_foreign_key table=%s column=%s references=%s.%s", requirement.Table, requirement.Column, requirement.ReferencedTable, requirement.ReferencedColumn)
	}
	return nil
}

func loadUniqueIndexes(ctx context.Context, db *gorm.DB, dialect Dialect, table string) ([]uniqueIndex, error) {
	var query string
	switch dialect {
	case PostgreSQL:
		query = `
			SELECT idx.relname, string_agg(att.attname, ',' ORDER BY key_columns.ordinality)
			FROM pg_class tbl
			JOIN pg_namespace ns ON ns.oid = tbl.relnamespace
			JOIN pg_index ind ON ind.indrelid = tbl.oid
			JOIN pg_class idx ON idx.oid = ind.indexrelid
			CROSS JOIN LATERAL unnest(ind.indkey) WITH ORDINALITY AS key_columns(attnum, ordinality)
			JOIN pg_attribute att ON att.attrelid = tbl.oid AND att.attnum = key_columns.attnum
			WHERE ns.nspname = current_schema() AND tbl.relname = ? AND ind.indisunique AND NOT ind.indisprimary
			GROUP BY idx.relname`
	case MySQL:
		query = `
			SELECT index_name, GROUP_CONCAT(column_name ORDER BY seq_in_index SEPARATOR ',')
			FROM information_schema.statistics
			WHERE table_schema = DATABASE() AND table_name = ? AND non_unique = 0 AND index_name <> 'PRIMARY'
			GROUP BY index_name`
	default:
		return nil, fmt.Errorf("unsupported database dialect %q", dialect)
	}

	rows, err := db.WithContext(ctx).Raw(query, table).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []uniqueIndex
	for rows.Next() {
		var name, columns string
		if err := rows.Scan(&name, &columns); err != nil {
			return nil, err
		}
		indexes = append(indexes, uniqueIndex{Name: name, Columns: strings.Split(columns, ",")})
	}
	return indexes, rows.Err()
}

func loadForeignKeys(ctx context.Context, db *gorm.DB, dialect Dialect, table string) ([]foreignKey, error) {
	var query string
	switch dialect {
	case PostgreSQL:
		query = `
			SELECT tc.constraint_name, kcu.column_name, ccu.table_name, ccu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
			  ON kcu.constraint_name = tc.constraint_name AND kcu.constraint_schema = tc.constraint_schema
			JOIN information_schema.constraint_column_usage ccu
			  ON ccu.constraint_name = tc.constraint_name AND ccu.constraint_schema = tc.constraint_schema
			WHERE tc.constraint_schema = current_schema() AND tc.table_name = ? AND tc.constraint_type = 'FOREIGN KEY'`
	case MySQL:
		query = `
			SELECT constraint_name, column_name, referenced_table_name, referenced_column_name
			FROM information_schema.key_column_usage
			WHERE table_schema = DATABASE() AND table_name = ? AND referenced_table_name IS NOT NULL`
	default:
		return nil, fmt.Errorf("unsupported database dialect %q", dialect)
	}

	rows, err := db.WithContext(ctx).Raw(query, table).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []foreignKey
	for rows.Next() {
		var key foreignKey
		if err := rows.Scan(&key.Name, &key.Column, &key.ReferencedTable, &key.ReferencedColumn); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func containsUniqueIndex(indexes []uniqueIndex, columns []string) bool {
	for _, index := range indexes {
		if sameStringSet(index.Columns, columns) {
			return true
		}
	}
	return false
}

func containsForeignKey(keys []foreignKey, requirement foreignKeyRequirement) bool {
	for _, key := range keys {
		if key.Column == requirement.Column && key.ReferencedTable == requirement.ReferencedTable && key.ReferencedColumn == requirement.ReferencedColumn {
			return true
		}
	}
	return false
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]string(nil), left...)
	rightCopy := append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	for i := range leftCopy {
		if leftCopy[i] != rightCopy[i] {
			return false
		}
	}
	return true
}

func quoteIdentifier(dialect Dialect, identifier string) string {
	if dialect == MySQL {
		return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
	}
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func quoteIdentifiers(dialect Dialect, identifiers []string) string {
	quoted := make([]string, 0, len(identifiers))
	for _, identifier := range identifiers {
		quoted = append(quoted, quoteIdentifier(dialect, identifier))
	}
	return strings.Join(quoted, ", ")
}
