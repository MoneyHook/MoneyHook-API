package model

import (
	"gorm.io/gorm"
)

type GetFixed struct {
	gorm.Model
	MonthlyTransactionId     string
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionSign   int
	MonthlyTransactionDate   int
	CategoryId               string
	CategoryName             string
	SubCategoryId            string
	SubCategoryName          string
	PaymentId                string
}

type GetDeletedFixed struct {
	gorm.Model
	MonthlyTransactionId     string
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionDate   int
	CategoryName             string
	SubCategoryName          string
}

type AddFixed struct {
	UserId                   string
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionDate   int
	CategoryId               string
	SubCategoryId            string
	SubCategoryName          string
	PaymentId                string
}

type EditFixed struct {
	UserId                   string
	MonthlyTransactionId     string
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionDate   int
	CategoryId               string
	SubCategoryId            string
	SubCategoryName          string
	IncludeFlg               bool
	PaymentId                string
}

type DeleteFixed struct {
	UserId               string
	MonthlyTransactionId string
}
