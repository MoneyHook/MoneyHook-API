package model

import "gorm.io/gorm"

type Timeline struct {
	gorm.Model
	TransactionId     int
	TransactionName   string
	TransactionAmount int
	TransactionSign   int
	TransactionDate   string
	CategoryId        int
	CategoryName      string
	SubCategoryId     int
	SubCategoryName   string
	FixedFlg          bool
}
