package model

import (
	"time"

	"gorm.io/gorm"
)

type JobMonthlyTransaction struct {
	gorm.Model
	MonthlyTransactionId     string
	UserNo                   string
	MonthlyTransactionName   string
	MonthlyTransactionAmount int
	MonthlyTransactionDate   int
	CategoryId               string
	SubCategoryId            string
	PaymentId                string
}

type JobTransaction struct {
	UserNo            string
	TransactionName   string
	TransactionAmount int
	TransactionDate   time.Time
	CategoryId        string
	SubCategoryId     string
	FixedFlg          bool
	PaymentId         string
}
