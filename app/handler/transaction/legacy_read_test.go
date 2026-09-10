package transaction

import (
	"MoneyHook/MoneyHook-API/message"
	"MoneyHook/MoneyHook-API/model"
	paymentdomain "MoneyHook/MoneyHook-API/paymentresource"
	transactiondomain "MoneyHook/MoneyHook-API/transaction"
	"errors"
	"github.com/labstack/echo/v4"
	"strings"
	"testing"
)

var readFailure = errors.New("private database failure")

type failingLegacyReads struct{ transactiondomain.Store }

func (s failingLegacyReads) GetTimelineData(userId string, month string) (*[]model.Timeline, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetMonthlySpendingData(userId string, month string) (*[]model.MonthlySpendingData, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetTransactionData(userId string, transactionId string) (*model.TransactionData, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetMonthlyFixedData(userId string, month string, isSpending bool) (*[]model.MonthlyFixedData, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetHome(userId string, month string) (*[]model.HomeCategory, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetMonthlyVariableData(userId string, month string) (*[]model.MonthlyVariableData, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetTotalSpending(userId string, categoryId string, subCategoryId string, startMonth string, endMonth string) (*[]model.TotalSpendingData, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetGroupByPayment(userId string, month string) (*[]model.PaymentGroupTransaction, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetLastMonthGroupByPayment(userId string, month string) (*[]model.PaymentGroupTransaction, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetMonthlyWithdrawalAmount(userId string, paymentId string, startMonth string, endMonth string) (*model.MonthlyWithdrawalAmountList, error) {
	return nil, readFailure
}
func (s failingLegacyReads) GetFrequentTransactionName(userId string, limit int) (*[]model.FrequentTransactionName, error) {
	return nil, readFailure
}

type failingPayments struct{ paymentdomain.Store }

func (s failingPayments) GetPaymentResourceList(userID string) (*[]model.PaymentResource, error) {
	return nil, readFailure
}
func TestLegacyReadErrorsReturn500(t *testing.T) {
	h := New(failingLegacyReads{}, failingPayments{})
	for name, handler := range map[string]echo.HandlerFunc{
		"GetTimelineData":            h.GetTimelineData,
		"GetMonthlySpendingData":     h.GetMonthlySpendingData,
		"GetTransaction":             h.GetTransaction,
		"GetMonthlyFixedIncome":      h.GetMonthlyFixedIncome,
		"GetMonthlyFixedSpending":    h.GetMonthlyFixedSpending,
		"GetHome":                    h.GetHome,
		"GetMonthlyVariableData":     h.GetMonthlyVariableData,
		"GetTotalSpendingData":       h.GetTotalSpendingData,
		"GroupByPayment":             h.GroupByPayment,
		"GetMonthlyWithdrawalAmount": h.GetMonthlyWithdrawalAmount,
		"GetFrequentTransactionName": h.GetFrequentTransactionName,
	} {
		t.Run(name, func(t *testing.T) {
			rec := runLegacyHandler(t, handler, "")
			if rec.Code != 500 || strings.Contains(rec.Body.String(), readFailure.Error()) {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

type missingLegacyTransaction struct{ transactiondomain.Store }

func (s missingLegacyTransaction) GetTransactionData(userID, transactionID string) (*model.TransactionData, error) {
	return nil, nil
}
func TestLegacyMissingTransactionStillReturns404(t *testing.T) {
	message.Read()
	rec := runLegacyHandler(t, New(missingLegacyTransaction{}, nil).GetTransaction, "")
	if rec.Code != 404 {
		t.Fatalf("status=%d", rec.Code)
	}
}

type priorMonthFailure struct{ failingLegacyReads }

func (s priorMonthFailure) GetGroupByPayment(userID, month string) (*[]model.PaymentGroupTransaction, error) {
	return &[]model.PaymentGroupTransaction{}, nil
}

type availablePayments struct{ paymentdomain.Store }

func (s availablePayments) GetPaymentResourceList(userID string) (*[]model.PaymentResource, error) {
	return &[]model.PaymentResource{{PaymentId: "1", PaymentDate: 27, ClosingDate: 31}}, nil
}
func TestLegacySecondaryReadFailuresReturn500(t *testing.T) {
	group := New(priorMonthFailure{}, nil)
	withdrawal := New(failingLegacyReads{}, availablePayments{})
	for name, handler := range map[string]echo.HandlerFunc{"previous month": group.GroupByPayment, "withdrawal aggregation": withdrawal.GetMonthlyWithdrawalAmount} {
		t.Run(name, func(t *testing.T) {
			rec := runLegacyHandler(t, handler, "")
			if rec.Code != 500 {
				t.Fatalf("status=%d", rec.Code)
			}
		})
	}
}
