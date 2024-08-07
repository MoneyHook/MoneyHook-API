package transaction

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetTimelineData(userId int, month string) *[]model.Timeline
	GetMonthlySpendingData(userId int, month string) *[]model.MonthlySpendingData
	GetTransactionData(userId int, transactionId int) *model.TransactionData
	GetMonthlyFixedData(userId int, month string, isSpending bool) *[]model.MonthlyFixedData
	GetHome(userId int, month string) *[]model.HomeCategory
	GetMonthlyVariableData(userId int, month string) *[]model.MonthlyVariableData
	GetTotalSpending(userId int, categoryId string, subCategoryId string, startMonth string, endMonth string) *[]model.TotalSpendingData
	GetGroupByPayment(userId int, month string) *[]model.PaymentGroupTransaction
	GetLastMonthGroupByPayment(userId int, month string) *[]model.PaymentGroupTransaction
	GetMonthlyWithdrawalAmount(userId int, paymentId int, startMonth string, endMonth string) *model.MonthlyWithdrawalAmountList
	GetFrequentTransactionName(userId int) *[]model.FrequentTransactionName
	AddTransaction(*model.AddTransaction) error
	AddTransactionList(*model.AddTransactionList) error
	EditTransaction(*model.EditTransaction) error
	DeleteTransaction(*model.DeleteTransaction) error
}
