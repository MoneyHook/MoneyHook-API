package transaction

import (
	"MoneyHook/MoneyHook-API/model"
	"errors"
)

var (
	ErrNotFound        = errors.New("transaction not found")
	ErrInvalidRelation = errors.New("invalid transaction relation")
)

type Store interface {
	GetTimelineData(userId string, month string) *[]model.Timeline
	GetMonthlySpendingData(userId string, month string) *[]model.MonthlySpendingData
	GetTransactionData(userId string, transactionId string) *model.TransactionData
	GetMonthlyFixedData(userId string, month string, isSpending bool) *[]model.MonthlyFixedData
	GetHome(userId string, month string) *[]model.HomeCategory
	GetMonthlyVariableData(userId string, month string) *[]model.MonthlyVariableData
	GetTotalSpending(userId string, categoryId string, subCategoryId string, startMonth string, endMonth string) *[]model.TotalSpendingData
	GetGroupByPayment(userId string, month string) *[]model.PaymentGroupTransaction
	GetLastMonthGroupByPayment(userId string, month string) *[]model.PaymentGroupTransaction
	GetMonthlyWithdrawalAmount(userId string, paymentId string, startMonth string, endMonth string) *model.MonthlyWithdrawalAmountList
	GetFrequentTransactionName(userId string) *[]model.FrequentTransactionName
	AddTransaction(*model.AddTransaction) error
	AddTransactionList(*model.AddTransactionList) error
	EditTransaction(*model.EditTransaction) error
	DeleteTransaction(*model.DeleteTransaction) error
	GetV1Transaction(userId string, transactionId string) (*model.V1Transaction, error)
	CreateV1Transaction(transaction *model.V1TransactionWrite) (*model.V1Transaction, error)
	UpdateV1Transaction(transaction *model.V1TransactionWrite) (*model.V1Transaction, string, error)
	DeleteV1Transaction(userId string, transactionId string) error
	GetV1AnalyticsTransactions(userId string, startDate string, endDate string) ([]model.V1AnalyticsTransaction, error)
}
