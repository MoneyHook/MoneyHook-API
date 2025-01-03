package db

import (
	"log"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
)

func New() *gorm.DB {
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

func NewMysql() *gorm.DB {

	log.Printf("Start MySQL Database Setup")
	db, err := gorm.Open(mysql.Open(getMySqlConfig()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Printf("Finish MySQL Database Setup")

	return db
}

func NewPostgres() *gorm.DB {

	log.Printf("Start PostgreSQL Database Setup")
	db, err := gorm.Open(postgres.Open(getPostgresConfig()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Printf("Finish PostgreSQL Database Setup")

	return db
}
