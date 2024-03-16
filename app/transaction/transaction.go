package transaction

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetTimelineData(userId int, month string) *[]model.Timeline
	GetMonthlySpendingData(userId int, month string) *[]model.MonthlySpendingData
	GetTransactionData(userId int, transactionId int) *model.TransactionData
}
