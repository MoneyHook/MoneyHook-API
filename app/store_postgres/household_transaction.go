package store_postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Retry only transactions the database has definitively rolled back. Callers
// keep all side effects inside fn; generated ids/results are overwritten on retry.
func householdTransaction(db *gorm.DB, fn func(*gorm.DB) error) error {
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = db.Transaction(fn)
		var pg *pgconn.PgError
		if !errors.As(err, &pg) || (pg.Code != "40001" && pg.Code != "40P01") {
			return err
		}
		timer := time.NewTimer(time.Duration(attempt+1) * 25 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return err
}
