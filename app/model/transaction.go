package model

import (
	"time"

	"gorm.io/gorm"
)

type Timeline struct {
	gorm.Model
	TransactionId     string
	TransactionName   string
	TransactionAmount int
	TransactionSign   int
	TransactionDate   time.Time
	CategoryId        string
	CategoryName      string
	SubCategoryId     string
	SubCategoryName   string
	FixedFlg          bool
	PaymentId         string
	PaymentName       string
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
	CategoryId        string
	CategoryName      string
	SubCategoryId     string
	SubCategoryName   string
	FixedFlg          bool
}

type MonthlyFixedData struct {
	gorm.Model
	CategoryId          string
	CategoryName        string
	TotalCategoryAmount int
	SubCategoryId       string
	SubCategoryName     string
	TransactionId       string
	TransactionName     string
	TransactionAmount   int
	TransactionDate     time.Time
	FixedFlg            bool
	PaymentId           string
	PaymentName         string
}

type HomeCategory struct {
	CategoryName           string
	CategoryTotalAmount    int
	SubCategoryName        string
	SubCategoryTotalAmount int
}

type MonthlyVariableData struct {
	CategoryId             string
	CategoryName           string
	CategoryTotalAmount    int
	SubCategoryId          string
	SubCategoryName        string
	SubCategoryTotalAmount int
	TransactionId          string
	TransactionName        string
	TransactionAmount      int
	TransactionDate        time.Time
	PaymentId              string
	PaymentName            string
}

type TotalSpendingData struct {
	CategoryName           string
	CategoryTotalAmount    int
	SubCategoryId          string
	SubCategoryName        string
	SubCategoryTotalAmount int
	TransactionId          string
	TransactionName        string
	TransactionAmount      int
	TransactionDate        time.Time
}

type PaymentGroupTransaction struct {
	PaymentId         string
	PaymentName       string
	PaymentAmount     int
	PaymentTypeId     string
	PaymentTypeName   string
	IsPaymentDueLater bool
	TransactionId     string
	TransactionName   string
	TransactionAmount int
	TransactionDate   time.Time
	CategoryId        string
	CategoryName      string
	SubCategoryId     string
	SubCategoryName   string
	FixedFlg          bool
}

type MonthlyWithdrawalAmountList struct {
	PaymentId            string
	PaymentName          string
	PaymentDate          int
	AggregationStartDate string
	AggregationEndDate   string
	WithdrawalAmount     int
}

type FrequentTransactionName struct {
	TransactionName string
	CategoryId      string
	SubCategoryId   string
	FixedFlg        bool
	PaymentId       string
	CategoryName    string
	SubCategoryName string
	RowNum          int
}

type AddTransaction struct {
	UserId            string
	TransactionDate   string
	TransactionAmount int
	TransactionName   string
	CategoryId        string
	SubCategoryId     string
	SubCategoryName   string
	FixedFlg          bool
	PaymentId         string
}

type AddTransactionList struct {
	UserId          string
	TransactionList []AddTransaction
}

type EditTransaction struct {
	TransactionId     string
	UserId            string
	TransactionDate   string
	TransactionAmount int
	TransactionName   string
	CategoryId        string
	SubCategoryId     string
	SubCategoryName   string
	FixedFlg          bool
	PaymentId         string
}

type DeleteTransaction struct {
	UserId        string
	TransactionId string
}
