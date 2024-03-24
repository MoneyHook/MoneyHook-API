package model

import (
	"gorm.io/gorm"
)

type GetFixed struct {
	gorm.Model
	MonthlyTransactionId     int
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionSign   int
	MonthlyTransactionDate   int
	CategoryId               int
	CategoryName             string
	SubCategoryId            int
	SubCategoryName          string
}

type GetDeletedFixed struct {
	gorm.Model
	MonthlyTransactionId     int
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionDate   int
	CategoryName             string
	SubCategoryName          string
}
