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

type AddFixed struct {
	UserId                   int
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionDate   int
	CategoryId               int
	SubCategoryId            int
	SubCategoryName          string
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
}

type DeleteFixed struct {
	UserId               int
	MonthlyTransactionId int
}
