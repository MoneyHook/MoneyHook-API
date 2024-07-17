package request

import (
	"MoneyHook/MoneyHook-API/model"

	"github.com/labstack/echo/v4"
)

type AddPaymentRequest struct {
	PaymentName   string `json:"payment_name"`
	PaymentTypeId *int   `json:"payment_type_id"`
	PaymentDate   *int   `json:"payment_date"`
	ClosingDate   *int   `json:"closing_date"`
}

func (r *AddPaymentRequest) Bind(c echo.Context, u *model.AddPaymentResource) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	u.PaymentName = r.PaymentName
	u.PaymentTypeId = 1
	if r.PaymentTypeId != nil {
		u.PaymentTypeId = *r.PaymentTypeId
		u.PaymentDate = r.PaymentDate
		u.ClosingDate = r.ClosingDate
	}

	return nil
}

type EditPaymentRequest struct {
	PaymentId     int    `json:"payment_id"`
	PaymentName   string `json:"payment_name"`
	PaymentTypeId *int   `json:"payment_type_id"`
	PaymentDate   *int   `json:"payment_date"`
	ClosingDate   *int   `json:"closing_date"`
}

func (r *EditPaymentRequest) Bind(c echo.Context, u *model.EditPaymentResource) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	// TODO バリデーション
	// if err := c.Validate(r); err != nil {
	// 	return err
	// }

	u.PaymentId = r.PaymentId
	u.PaymentName = r.PaymentName
	u.PaymentTypeId = 1
	if r.PaymentTypeId != nil {
		u.PaymentTypeId = *r.PaymentTypeId
		u.PaymentDate = r.PaymentDate
		u.ClosingDate = r.ClosingDate
	}

	return nil
}
