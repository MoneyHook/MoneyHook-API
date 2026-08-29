package migration

import (
	"context"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

func removeLegacyAuthSchema(ctx context.Context, db *gorm.DB, dialect Dialect) error {
	if db.WithContext(ctx).Migrator().HasTable("user_token") {
		if err := db.WithContext(ctx).Migrator().DropTable("user_token"); err != nil {
			return fmt.Errorf("drop user_token table: %w", err)
		}
		log.Printf("event=schema_migration_change action=drop_table table=user_token")
	}
	if !db.WithContext(ctx).Migrator().HasTable("users") {
		return nil
	}

	columns, err := loadColumns(ctx, db, "users")
	if err != nil {
		return err
	}
	for column := range columns {
		if !strings.EqualFold(column, "email") && !strings.EqualFold(column, "password") {
			continue
		}
		query := fmt.Sprintf(
			"ALTER TABLE %s DROP COLUMN %s",
			quoteIdentifier(dialect, "users"),
			quoteIdentifier(dialect, column),
		)
		if err := db.WithContext(ctx).Exec(query).Error; err != nil {
			return fmt.Errorf("drop users.%s: %w", column, err)
		}
		log.Printf("event=schema_migration_change action=drop_column table=users column=%s", column)
	}

	var duplicate struct {
		UserID string `gorm:"column:user_id"`
		Count  int64  `gorm:"column:duplicate_count"`
	}
	if err := db.WithContext(ctx).Raw(`
		SELECT user_id, COUNT(*) AS duplicate_count
		FROM users
		GROUP BY user_id
		HAVING COUNT(*) > 1
		LIMIT 1`).Scan(&duplicate).Error; err != nil {
		return fmt.Errorf("inspect duplicate Firebase identities: %w", err)
	}
	if duplicate.Count > 1 {
		return fmt.Errorf("users.user_id %q is duplicated %d times", duplicate.UserID, duplicate.Count)
	}
	return nil
}
