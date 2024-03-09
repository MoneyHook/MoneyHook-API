package model

import (
	"time"

	"gorm.io/gorm"
)

type Timeline struct {
	gorm.Model
	TransactionId     int
	TransactionName   string
	TransactionAmount int
	TransactionSign   int
	TransactionDate   time.Time
	CategoryId        int
	CategoryName      string
	SubCategoryId     int
	SubCategoryName   string
	FixedFlg          bool
}
