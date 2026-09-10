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
	GetTimelineData(userId string, month string) (*[]model.Timeline, error)
	GetMonthlySpendingData(userId string, month string) (*[]model.MonthlySpendingData, error)
	GetTransactionData(userId string, transactionId string) (*model.TransactionData, error)
	GetMonthlyFixedData(userId string, month string, isSpending bool) (*[]model.MonthlyFixedData, error)
	GetHome(userId string, month string) (*[]model.HomeCategory, error)
	GetMonthlyVariableData(userId string, month string) (*[]model.MonthlyVariableData, error)
	GetTotalSpending(userId string, categoryId string, subCategoryId string, startMonth string, endMonth string) (*[]model.TotalSpendingData, error)
	GetGroupByPayment(userId string, month string) (*[]model.PaymentGroupTransaction, error)
	GetLastMonthGroupByPayment(userId string, month string) (*[]model.PaymentGroupTransaction, error)
	GetMonthlyWithdrawalAmount(userId string, paymentId string, startMonth string, endMonth string) (*model.MonthlyWithdrawalAmountList, error)
	GetFrequentTransactionName(userId string, limit int) (*[]model.FrequentTransactionName, error)
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
