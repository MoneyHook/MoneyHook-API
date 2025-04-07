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
	CategoryId        int
	CategoryName      string
	SubCategoryId     int
	SubCategoryName   string
	FixedFlg          bool
	PaymentId         int
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
	CategoryId        int
	CategoryName      string
	SubCategoryId     int
	SubCategoryName   string
	FixedFlg          bool
}

type MonthlyFixedData struct {
	gorm.Model
	CategoryId          int
	CategoryName        string
	TotalCategoryAmount int
	SubCategoryId       int
	SubCategoryName     string
	TransactionId       string
	TransactionName     string
	TransactionAmount   int
	TransactionDate     time.Time
	FixedFlg            bool
	PaymentId           int
	PaymentName         string
}

type HomeCategory struct {
	CategoryName           string
	CategoryTotalAmount    int
	SubCategoryName        string
	SubCategoryTotalAmount int
}

type MonthlyVariableData struct {
	CategoryId             int
	CategoryName           string
	CategoryTotalAmount    int
	SubCategoryId          int
	SubCategoryName        string
	SubCategoryTotalAmount int
	TransactionId          string
	TransactionName        string
	TransactionAmount      int
	TransactionDate        time.Time
	PaymentId              int
	PaymentName            string
}

type TotalSpendingData struct {
	CategoryName           string
	CategoryTotalAmount    int
	SubCategoryId          int
	SubCategoryName        string
	SubCategoryTotalAmount int
	TransactionId          string
	TransactionName        string
	TransactionAmount      int
	TransactionDate        time.Time
}

type PaymentGroupTransaction struct {
	PaymentId         int
	PaymentName       string
	PaymentAmount     int
	PaymentTypeId     int
	PaymentTypeName   string
	IsPaymentDueLater bool
	TransactionId     string
	TransactionName   string
	TransactionAmount int
	TransactionDate   time.Time
	CategoryId        int
	CategoryName      string
	SubCategoryId     int
	SubCategoryName   string
	FixedFlg          bool
}

type MonthlyWithdrawalAmountList struct {
	PaymentId            int
	PaymentName          string
	PaymentDate          int
	AggregationStartDate string
	AggregationEndDate   string
	WithdrawalAmount     int
}

type FrequentTransactionName struct {
	TransactionName string
	CategoryId      int
	SubCategoryId   int
	FixedFlg        bool
	PaymentId       int
	CategoryName    string
	SubCategoryName string
	RowNum          int
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
	PaymentId         int
}

type AddTransactionList struct {
	UserId          int
	TransactionList []AddTransaction
}

type EditTransaction struct {
	TransactionId     string
	UserId            int
	TransactionDate   string
	TransactionAmount int
	TransactionName   string
	CategoryId        int
	SubCategoryId     int
	SubCategoryName   string
	FixedFlg          bool
	PaymentId         int
}

type DeleteTransaction struct {
	UserId        int
	TransactionId string
}
