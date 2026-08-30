package db

import (
	category "MoneyHook/MoneyHook-API/category"
	common "MoneyHook/MoneyHook-API/common"
	dbmigration "MoneyHook/MoneyHook-API/db/migration"
	fixed "MoneyHook/MoneyHook-API/fixed"
	job "MoneyHook/MoneyHook-API/job"
	paymentresource "MoneyHook/MoneyHook-API/paymentresource"
	"MoneyHook/MoneyHook-API/store_mysql"
	"MoneyHook/MoneyHook-API/store_postgres"
	subcategory "MoneyHook/MoneyHook-API/subcategory"
	transaction "MoneyHook/MoneyHook-API/transaction"
	user "MoneyHook/MoneyHook-API/user"

	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Store struct {
	UserStore            user.Store
	TransactionStore     transaction.Store
	FixedStore           fixed.Store
	CategoryStore        category.Store
	SubCategoryStore     subcategory.Store
	PaymentResourceStore paymentresource.Store
	JobStore             job.Store
}

type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
)

func New() (*Store, error) {
	enableSeedData, err := SeedDataEnabledFromEnvironment()
	if err != nil {
		return nil, err
	}
	dbType := DatabaseType(strings.ToLower(common.GetEnv("DATABASE_TYPE", "MySQL")))

	switch dbType {
	case MySQL:
		return newMysql(enableSeedData)
	case PostgreSQL:
		return newPostgres(enableSeedData)
	default:
		return nil, fmt.Errorf("unsupported DATABASE_TYPE %q: set MySQL or PostgreSQL", dbType)
	}
}

func NewMysql() (*Store, error) {
	enableSeedData, err := SeedDataEnabledFromEnvironment()
	if err != nil {
		return nil, err
	}
	return newMysql(enableSeedData)
}

func newMysql(enableSeedData bool) (*Store, error) {

	log.Printf("Start MySQL Database Setup")
	db, err := openWithRetry("MySQL", func() gorm.Dialector {
		return mysql.Open(getMySqlConfig())
	})
	if err != nil {
		return nil, err
	}
	if err := dbmigration.Run(
		context.Background(),
		db,
		dbmigration.MySQL,
		common.GetEnv("MYSQL_DATABASE", ""),
		dbmigration.Options{EnableSeedData: enableSeedData},
	); err != nil {
		return nil, err
	}
	log.Printf("Finish MySQL Database Setup")

	us := store_mysql.NewUserStore(db)
	ts := store_mysql.NewTransactionStore(db)
	fs := store_mysql.NewFixedStore(db)
	cs := store_mysql.NewCategoryStore(db)
	scs := store_mysql.NewSubCategoryStore(db)
	pr := store_mysql.NewPaymentResourceStore(db)
	job := store_mysql.NewJobStore(db)

	return &Store{UserStore: us, TransactionStore: ts, FixedStore: fs, CategoryStore: cs, SubCategoryStore: scs, PaymentResourceStore: pr, JobStore: job}, nil
}

func NewPostgres() (*Store, error) {
	enableSeedData, err := SeedDataEnabledFromEnvironment()
	if err != nil {
		return nil, err
	}
	return newPostgres(enableSeedData)
}

func newPostgres(enableSeedData bool) (*Store, error) {

	log.Printf("Start PostgreSQL Database Setup")
	databaseName := common.GetEnv("POSTGRES_DATABASE", "")
	if err := ensurePostgresDatabase(context.Background(), databaseName); err != nil {
		return nil, err
	}
	db, err := openWithRetry("PostgreSQL", func() gorm.Dialector {
		return postgres.Open(getPostgresConfig())
	})
	if err != nil {
		return nil, err
	}
	if err := dbmigration.Run(
		context.Background(),
		db,
		dbmigration.PostgreSQL,
		databaseName,
		dbmigration.Options{EnableSeedData: enableSeedData},
	); err != nil {
		return nil, err
	}
	log.Printf("Finish PostgreSQL Database Setup")

	us := store_postgres.NewUserStore(db)
	ts := store_postgres.NewTransactionStore(db)
	fs := store_postgres.NewFixedStore(db)
	cs := store_postgres.NewCategoryStore(db)
	scs := store_postgres.NewSubCategoryStore(db)
	pr := store_postgres.NewPaymentResourceStore(db)
	job := store_postgres.NewJobStore(db)

	return &Store{UserStore: us, TransactionStore: ts, FixedStore: fs, CategoryStore: cs, SubCategoryStore: scs, PaymentResourceStore: pr, JobStore: job}, nil
}

func ensurePostgresDatabase(ctx context.Context, databaseName string) error {
	if strings.TrimSpace(databaseName) == "" {
		return fmt.Errorf("POSTGRES_DATABASE must not be empty")
	}

	adminDB, err := openWithRetry("PostgreSQL admin database", func() gorm.Dialector {
		return postgres.Open(getPostgresAdminConfig())
	})
	if err != nil {
		return err
	}
	sqlDB, err := adminDB.DB()
	if err != nil {
		return fmt.Errorf("get PostgreSQL admin database handle: %w", err)
	}
	defer sqlDB.Close()

	var exists bool
	if err := adminDB.WithContext(ctx).
		Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", databaseName).
		Scan(&exists).Error; err != nil {
		return fmt.Errorf("check PostgreSQL database %q: %w", databaseName, err)
	}
	if exists {
		log.Printf("event=database_creation status=already_exists dialect=postgresql database=%s", databaseName)
		return nil
	}

	query := fmt.Sprintf("CREATE DATABASE %s", quotePostgresIdentifier(databaseName))
	if err := adminDB.WithContext(ctx).Exec(query).Error; err != nil {
		if checkErr := adminDB.WithContext(ctx).
			Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", databaseName).
			Scan(&exists).Error; checkErr == nil && exists {
			log.Printf("event=database_creation status=already_exists dialect=postgresql database=%s", databaseName)
			return nil
		}
		return fmt.Errorf("create PostgreSQL database %q: %w", databaseName, err)
	}

	log.Printf("event=database_creation status=created dialect=postgresql database=%s", databaseName)
	return nil
}

func quotePostgresIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func SeedDataEnabledFromEnvironment() (bool, error) {
	value, exists := os.LookupEnv("ENABLE_SEED_DATA")
	return parseSeedDataEnabled(value, exists)
}

func parseSeedDataEnabled(value string, exists bool) (bool, error) {
	value = strings.TrimSpace(value)
	if !exists || value == "" {
		return false, nil
	}

	enabled, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("invalid ENABLE_SEED_DATA value %q: expected a boolean accepted by strconv.ParseBool", value)
	}
	return enabled, nil
}

func openWithRetry(databaseName string, dialector func() gorm.Dialector) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	delay := time.Second
	var lastErr error
	for attempt := 1; ; attempt++ {
		db, err := gorm.Open(dialector(), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			IgnoreRelationshipsWhenMigrating:         true,
		})
		if err == nil {
			return db, nil
		}
		lastErr = err
		if db != nil {
			if sqlDB, sqlErr := db.DB(); sqlErr == nil {
				_ = sqlDB.Close()
			}
		}

		log.Printf("event=database_connection_retry database=%s attempt=%d delay_ms=%d error=%q", databaseName, attempt, delay.Milliseconds(), err)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("connect to %s within 60 seconds: %w", databaseName, lastErr)
		case <-timer.C:
		}
		if delay < 5*time.Second {
			delay *= 2
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
		}
	}
}
