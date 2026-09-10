package transaction

import (
	"MoneyHook/MoneyHook-API/model"
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/labstack/echo/v4"
)

type Transaction struct {
	TransactionDate   string `json:"transaction_date"`
	TransactionAmount *int   `json:"transaction_amount"`
	TransactionSign   int    `json:"transaction_sign"`
	TransactionName   string `json:"transaction_name"`
	CategoryId        string `json:"category_id"`
	SubCategoryId     string `json:"sub_category_id"`
	SubCategoryName   string `json:"sub_category_name"`
	FixedFlg          *bool  `json:"fixed_flg"`
	PaymentId         string `json:"payment_id"`
}

func (r Transaction) validate() error {
	if !isISODate(r.TransactionDate) || strings.HasPrefix(r.TransactionDate, "0000-") {
		return errors.New("transaction_date must be a valid calendar date")
	}
	if r.TransactionAmount == nil || *r.TransactionAmount < 0 {
		return errors.New("transaction_amount must be a non-negative integer")
	}
	if r.TransactionSign != -1 && r.TransactionSign != 1 {
		return errors.New("transaction_sign must be -1 or 1")
	}
	if strings.TrimSpace(r.TransactionName) == "" || utf8.RuneCountInString(r.TransactionName) > 32 {
		return errors.New("transaction_name must contain 1 to 32 characters")
	}
	if !isLegacyID(r.CategoryId) || (r.SubCategoryId != "" && !isLegacyID(r.SubCategoryId)) || (r.PaymentId != "" && !isLegacyID(r.PaymentId)) {
		return errors.New("IDs must be positive database integers")
	}
	if utf8.RuneCountInString(r.SubCategoryName) > 16 || (r.SubCategoryId == "" && strings.TrimSpace(r.SubCategoryName) == "") {
		return errors.New("sub_category_name must contain 1 to 16 characters when sub_category_id is omitted")
	}
	if r.FixedFlg == nil {
		return errors.New("fixed_flg is required")
	}
	return nil
}

func isLegacyID(value string) bool {
	id, err := strconv.ParseInt(value, 10, 64)
	return err == nil && id > 0
}

func (r Transaction) toModel() model.AddTransaction {
	return model.AddTransaction{
		TransactionDate:   r.TransactionDate,
		TransactionAmount: *r.TransactionAmount * r.TransactionSign,
		TransactionName:   r.TransactionName,
		CategoryId:        r.CategoryId,
		SubCategoryId:     r.SubCategoryId,
		SubCategoryName:   r.SubCategoryName,
		FixedFlg:          *r.FixedFlg,
		PaymentId:         r.PaymentId,
	}
}

type AddTransactionRequest struct {
	Transaction Transaction `json:"transaction"`
}

func (r *AddTransactionRequest) Bind(c echo.Context, u *model.AddTransaction) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := r.Transaction.validate(); err != nil {
		return err
	}
	input := r.Transaction.toModel()
	input.UserId = u.UserId
	*u = input
	return nil
}

type AddTransactionListRequest struct {
	TransactionList []Transaction `json:"transaction_list"`
}

func (r *AddTransactionListRequest) Bind(c echo.Context, u *model.AddTransactionList) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	if len(r.TransactionList) == 0 {
		return errors.New("transaction_list must not be empty")
	}
	// Validate every row before producing input for persistence.
	for _, transaction := range r.TransactionList {
		if err := transaction.validate(); err != nil {
			return err
		}
	}
	transactions := make([]model.AddTransaction, 0, len(r.TransactionList))
	for _, transaction := range r.TransactionList {
		transactions = append(transactions, transaction.toModel())
	}
	u.TransactionList = transactions
	return nil
}

type EditTransactionRequest struct {
	Transaction struct {
		Transaction
		TransactionId string `json:"transaction_id"`
	} `json:"transaction"`
}

func (r *EditTransactionRequest) Bind(c echo.Context, u *model.EditTransaction) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := r.Transaction.validate(); err != nil {
		return err
	}
	if !isLegacyID(r.Transaction.TransactionId) {
		return errors.New("transaction_id must be a positive database integer")
	}
	input := r.Transaction.toModel()
	*u = model.EditTransaction{
		UserId:            u.UserId,
		TransactionId:     r.Transaction.TransactionId,
		TransactionDate:   input.TransactionDate,
		TransactionAmount: input.TransactionAmount,
		TransactionName:   input.TransactionName,
		CategoryId:        input.CategoryId,
		SubCategoryId:     input.SubCategoryId,
		SubCategoryName:   input.SubCategoryName,
		FixedFlg:          input.FixedFlg,
		PaymentId:         input.PaymentId,
	}
	return nil
}
