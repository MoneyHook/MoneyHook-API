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

type MonthlySpendingData struct {
	TotalAmount int
	Month       string
}

type TransactionData struct {
	gorm.Model
	TransactionDate   time.Time
	TransactionName   string
	TransactionAmount int
	CategoryId        int
	CategoryName      string
	SubCategoryId     int
	SubCategoryName   string
	FixedFlg          bool
}

type MonthlyFixedIncome struct{
	gorm.Model
	CategoryId int 
	CategoryName string
    TotalCategoryAmount int
	TransactionCategoryId int
	TransactionName string
	TransactionAmount int
}