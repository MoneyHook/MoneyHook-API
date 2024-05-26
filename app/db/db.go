package db

import (
	"log"

	"gorm.io/driver/mysql"
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
