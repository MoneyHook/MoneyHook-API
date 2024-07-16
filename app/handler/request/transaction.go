package request

import (
	"MoneyHook/MoneyHook-API/model"

	"github.com/labstack/echo/v4"
)

type Transaction struct {
	TransactionDate   string `json:"transaction_date" validate:"required"`
	TransactionAmount int    `json:"transaction_amount"  validate:"required"`
	TransactionSign   int    `json:"transaction_sign"  validate:"required"`
	TransactionName   string `json:"transaction_name"  validate:"required"`
	CategoryId        int    `json:"category_id"  validate:"required"`
	SubCategoryId     int    `json:"sub_category_id"`
	SubCategoryName   string `json:"sub_category_name"`
	FixedFlg          bool   `json:"fixed_flg"  validate:"required"`
	PaymentId         int    `json:"payment_id"`
}

type AddTransactionRequest struct {
	Transaction `json:"transaction"`
}

func (r *AddTransactionRequest) Bind(c echo.Context, u *model.AddTransaction) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	u.TransactionDate = r.Transaction.TransactionDate
	u.TransactionAmount = r.Transaction.TransactionAmount * r.Transaction.TransactionSign
	u.TransactionName = r.Transaction.TransactionName
	u.CategoryId = r.Transaction.CategoryId
	u.SubCategoryId = r.Transaction.SubCategoryId
	u.SubCategoryName = r.Transaction.SubCategoryName
	u.FixedFlg = r.Transaction.FixedFlg
	u.PaymentId = r.Transaction.PaymentId

	return nil
}

type AddTransactionListRequest struct {
	TransactionList []Transaction `json:"transaction_list" validate:"required,dive"`
}

func (r *AddTransactionListRequest) Bind(c echo.Context, u *model.AddTransactionList) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	for _, addTran := range r.TransactionList {
		u.TransactionList = append(u.TransactionList,
			model.AddTransaction{
				TransactionDate:   addTran.TransactionDate,
				TransactionAmount: addTran.TransactionAmount * addTran.TransactionSign,
				TransactionName:   addTran.TransactionName,
				CategoryId:        addTran.CategoryId,
				SubCategoryId:     addTran.SubCategoryId,
				SubCategoryName:   addTran.SubCategoryName,
				FixedFlg:          addTran.FixedFlg,
				PaymentId:         addTran.PaymentId})
	}

	return nil
}

type EditTransactionRequest struct {
	Transaction struct {
		TransactionId     int    `json:"transaction_id" validate:"required"`
		TransactionDate   string `json:"transaction_date" validate:"required"`
		TransactionAmount int    `json:"transaction_amount"  validate:"required"`
		TransactionSign   int    `json:"transaction_sign"  validate:"required"`
		TransactionName   string `json:"transaction_name"  validate:"required"`
		CategoryId        int    `json:"category_id"  validate:"required"`
		SubCategoryId     int    `json:"sub_category_id"`
		SubCategoryName   string `json:"sub_category_name"`
		FixedFlg          bool   `json:"fixed_flg"  validate:"required"`
		PaymentId         int    `json:"payment_id"`
	} `json:"transaction"`
}

func (r *EditTransactionRequest) Bind(c echo.Context, u *model.EditTransaction) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	u.TransactionId = r.Transaction.TransactionId
	u.TransactionDate = r.Transaction.TransactionDate
	u.TransactionAmount = r.Transaction.TransactionAmount * r.Transaction.TransactionSign
	u.TransactionName = r.Transaction.TransactionName
	u.CategoryId = r.Transaction.CategoryId
	u.SubCategoryId = r.Transaction.SubCategoryId
	u.SubCategoryName = r.Transaction.SubCategoryName
	u.FixedFlg = r.Transaction.FixedFlg
	u.PaymentId = r.Transaction.PaymentId

	return nil
}
