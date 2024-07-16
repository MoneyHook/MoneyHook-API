package request

import (
	"MoneyHook/MoneyHook-API/model"

	"github.com/labstack/echo/v4"
)

type AddFixedRequest struct {
	MonthlyTransaction struct {
		MonthlyTransactionName   string `json:"monthly_transaction_name"  validate:"required"`
		MonthlyTransactionAmount int    `json:"monthly_transaction_amount"  validate:"required"`
		MonthlyTransactionSign   int    `json:"monthly_transaction_sign"  validate:"required"`
		MonthlyTransactionDate   int    `json:"monthly_transaction_date" validate:"required"`
		CategoryId               int    `json:"category_id"  validate:"required"`
		SubCategoryId            int    `json:"sub_category_id"`
		SubCategoryName          string `json:"sub_category_name"`
		PaymentId                int    `json:"payment_id"`
	} `json:"monthly_transaction"`
}

func (r *AddFixedRequest) Bind(c echo.Context, u *model.AddFixed) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	u.MonthlyTransactionDate = r.MonthlyTransaction.MonthlyTransactionDate
	u.MonthlyTransactionAmount = r.MonthlyTransaction.MonthlyTransactionAmount * r.MonthlyTransaction.MonthlyTransactionSign
	u.MonthlyTransactionName = r.MonthlyTransaction.MonthlyTransactionName
	u.CategoryId = r.MonthlyTransaction.CategoryId
	u.SubCategoryId = r.MonthlyTransaction.SubCategoryId
	u.SubCategoryName = r.MonthlyTransaction.SubCategoryName
	u.PaymentId = r.MonthlyTransaction.PaymentId

	return nil
}

type EditFixedRequest struct {
	MonthlyTransaction struct {
		MonthlyTransactionId     int    `json:"monthly_transaction_id"  validate:"required"`
		MonthlyTransactionName   string `json:"monthly_transaction_name"  validate:"required"`
		MonthlyTransactionAmount int    `json:"monthly_transaction_amount"  validate:"required"`
		MonthlyTransactionSign   int    `json:"monthly_transaction_sign"  validate:"required"`
		MonthlyTransactionDate   int    `json:"monthly_transaction_date" validate:"required"`
		CategoryId               int    `json:"category_id"  validate:"required"`
		SubCategoryId            int    `json:"sub_category_id"`
		SubCategoryName          string `json:"sub_category_name"`
		IncludeFlg               bool   `json:"include_flg"`
		PaymentId                int    `json:"payment_id"`
	} `json:"monthly_transaction"`
}

func (r *EditFixedRequest) Bind(c echo.Context, u *model.EditFixed) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	u.MonthlyTransactionId = r.MonthlyTransaction.MonthlyTransactionId
	u.MonthlyTransactionDate = r.MonthlyTransaction.MonthlyTransactionDate
	u.MonthlyTransactionAmount = r.MonthlyTransaction.MonthlyTransactionAmount * r.MonthlyTransaction.MonthlyTransactionSign
	u.MonthlyTransactionName = r.MonthlyTransaction.MonthlyTransactionName
	u.CategoryId = r.MonthlyTransaction.CategoryId
	u.SubCategoryId = r.MonthlyTransaction.SubCategoryId
	u.SubCategoryName = r.MonthlyTransaction.SubCategoryName
	u.IncludeFlg = r.MonthlyTransaction.IncludeFlg
	u.PaymentId = r.MonthlyTransaction.PaymentId

	return nil
}
