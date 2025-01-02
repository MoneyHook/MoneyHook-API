package db

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New() *gorm.DB {

	log.Printf("Start Database Setup")
	db, err := gorm.Open(mysql.Open(getConfig()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Printf("Finish Database Setup")

	return db
}

func NewPostgres() *gorm.DB {

	log.Printf("Start Database Setup")
	db, err := gorm.Open(postgres.Open(getPostgresConfig()), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	log.Printf("Finish Database Setup")

	return db
}
