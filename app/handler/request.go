package handler

import (
	"MoneyHook/MoneyHook-API/model"

	"github.com/labstack/echo/v4"
)

type addTransactionRequest struct {
	Transaction struct {
		TransactionDate   string `json:"transaction_date" validate:"required"`
		TransactionAmount int    `json:"transaction_amount"  validate:"required"`
		TransactionSign   int    `json:"transaction_sign"  validate:"required"`
		TransactionName   string `json:"transaction_name"  validate:"required"`
		CategoryId        int    `json:"category_id"  validate:"required"`
		SubCategoryId     int    `json:"sub_category_id"`
		SubCategoryName   string `json:"sub_category_name"`
		FixedFlg          bool   `json:"fixed_flg"  validate:"required"`
	} `json:"transaction"`
}

func (r *addTransactionRequest) bind(c echo.Context, u *model.AddTransaction) error {
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

	return nil
}
