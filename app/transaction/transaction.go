package transaction

import "MoneyHook/MoneyHook-API/model"

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
}
