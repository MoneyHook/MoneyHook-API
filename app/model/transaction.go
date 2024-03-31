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

type MonthlyFixedData struct {
	gorm.Model
	CategoryName        string
	TotalCategoryAmount int
	TransactionName     string
	TransactionAmount   int
}

type HomeCategory struct {
	CategoryName           string
	CategoryTotalAmount    int
	SubCategoryName        string
	SubCategoryTotalAmount int
}

type MonthlyVariableData struct {
	CategoryName           string
	CategoryTotalAmount    int
	SubCategoryId          int
	SubCategoryName        string
	SubCategoryTotalAmount int
	TransactionId          int
	TransactionName        string
	TransactionAmount      int
}

type TotalSpendingData struct {
	CategoryName           string
	CategoryTotalAmount    int
	SubCategoryId          int
	SubCategoryName        string
	SubCategoryTotalAmount int
	TransactionId          int
	TransactionName        string
	TransactionAmount      int
	TransactionDate        time.Time
}

type AddTransaction struct {
	UserId            int
	TransactionDate   string
	TransactionAmount int
	TransactionName   string
	CategoryId        int
	SubCategoryId     int
	SubCategoryName   string
	FixedFlg          bool
}
