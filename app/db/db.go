package db

import (
	category "MoneyHook/MoneyHook-API/cagegory"
	fixed "MoneyHook/MoneyHook-API/fixed"
	payment_resource "MoneyHook/MoneyHook-API/payment_resource"
	"MoneyHook/MoneyHook-API/store_mysql"
	"MoneyHook/MoneyHook-API/store_postgres"
	sub_category "MoneyHook/MoneyHook-API/sub_cagegory"
	transaction "MoneyHook/MoneyHook-API/transaction"
	user "MoneyHook/MoneyHook-API/user"

	"log"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Store struct {
	UserStore user.Store
	TransactionStore transaction.Store
	FixedStore fixed.Store
	CategoryStore category.Store
	SubCategoryStore sub_category.Store
	PaymentResourceStore payment_resource.Store
}

type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
)

func New() *Store {
	dbType := DatabaseType(strings.ToLower(os.Getenv("DATABASE_TYPE")))

	switch dbType {
	case MySQL:
		return NewMysql()
	case PostgreSQL:
		return NewPostgres()
	default:
		log.Fatalf("Unsupported DATABASE_TYPE: '%s'. Please set 'MySQL' or 'PostgreSQL'", dbType)
		return nil
	}
}

func NewMysql() *Store {

	log.Printf("Start MySQL Database Setup")
	db, err := gorm.Open(mysql.Open(getMySqlConfig()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Printf("Finish MySQL Database Setup")

	us := store_mysql.NewUserStore(db)
	ts := store_mysql.NewTransactionStore(db)
	fs := store_mysql.NewFixedStore(db)
	cs := store_mysql.NewCategoryStore(db)
	scs := store_mysql.NewSubCategoryStore(db)
	pr := store_mysql.NewPaymentResourceStore(db)


	return &Store{UserStore: us, TransactionStore: ts, FixedStore: fs, CategoryStore: cs, SubCategoryStore: scs, PaymentResourceStore: pr}
}

func NewPostgres() *Store {

	log.Printf("Start PostgreSQL Database Setup")
	db, err := gorm.Open(postgres.Open(getPostgresConfig()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Printf("Finish PostgreSQL Database Setup")

	us := store_postgres.NewUserStore(db)
	ts := store_postgres.NewTransactionStore(db)
	fs := store_postgres.NewFixedStore(db)
	cs := store_postgres.NewCategoryStore(db)
	scs := store_postgres.NewSubCategoryStore(db)
	pr := store_postgres.NewPaymentResourceStore(db)

	return &Store{UserStore: us, TransactionStore: ts, FixedStore: fs, CategoryStore: cs, SubCategoryStore: scs, PaymentResourceStore: pr}
}
