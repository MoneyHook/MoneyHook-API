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

type editTransactionRequest struct {
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
	} `json:"transaction"`
}

func (r *editTransactionRequest) bind(c echo.Context, u *model.EditTransaction) error {
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

	return nil
}

type editSubCategoryRequest struct {
	SubCategoryId int  `json:"sub_category_id" validate:"required"`
	IsEnable      bool `json:"is_enable"  validate:"required"`
}

func (r *editSubCategoryRequest) bind(c echo.Context, u *model.EditSubCategoryModel) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	u.SubCategoryId = r.SubCategoryId
	u.IsEnable = r.IsEnable

	return nil
}

type addFixedRequest struct {
	MonthlyTransaction struct {
		MonthlyTransactionName   string `json:"monthly_transaction_name"  validate:"required"`
		MonthlyTransactionAmount int    `json:"monthly_transaction_amount"  validate:"required"`
		MonthlyTransactionSign   int    `json:"monthly_transaction_sign"  validate:"required"`
		MonthlyTransactionDate   int    `json:"monthly_transaction_date" validate:"required"`
		CategoryId               int    `json:"category_id"  validate:"required"`
		SubCategoryId            int    `json:"sub_category_id"`
		SubCategoryName          string `json:"sub_category_name"`
	} `json:"monthly_transaction"`
}

func (r *addFixedRequest) bind(c echo.Context, u *model.AddFixed) error {
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

	return nil
}

type editFixedRequest struct {
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
	} `json:"monthly_transaction"`
}

func (r *editFixedRequest) bind(c echo.Context, u *model.EditFixed) error {
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

	return nil
}
