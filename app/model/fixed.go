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
	PaymentId                int
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

type AddFixed struct {
	UserId                   int
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionDate   int
	CategoryId               int
	SubCategoryId            int
	SubCategoryName          string
	PaymentId                int
}

type EditFixed struct {
	UserId                   int
	MonthlyTransactionId     int
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionDate   int
	CategoryId               int
	SubCategoryId            int
	SubCategoryName          string
	IncludeFlg               bool
	PaymentId                int
}

type DeleteFixed struct {
	UserId               int
	MonthlyTransactionId int
}
