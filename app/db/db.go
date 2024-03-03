package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func New() *gorm.DB {

	db, err := gorm.Open(mysql.Open(getConfig()), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}
