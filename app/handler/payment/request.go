package payment

import (
	"MoneyHook/MoneyHook-API/model"

	"github.com/labstack/echo/v4"
)

type AddPaymentRequest struct {
	PaymentName   string  `json:"payment_name"`
	PaymentTypeId *string `json:"payment_type_id"`
	PaymentDate   *int    `json:"payment_date"`
	ClosingDate   *int    `json:"closing_date"`
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
	u.PaymentTypeId = "1"
	if r.PaymentTypeId != nil {
		u.PaymentTypeId = *r.PaymentTypeId
		u.PaymentDate = r.PaymentDate
		u.ClosingDate = r.ClosingDate
	}

	return nil
}

type EditPaymentRequest struct {
	PaymentId     string  `json:"payment_id"`
	PaymentName   string  `json:"payment_name"`
	PaymentTypeId *string `json:"payment_type_id"`
	PaymentDate   *int    `json:"payment_date"`
	ClosingDate   *int    `json:"closing_date"`
}

type ReorderPaymentResourcesRequest struct {
	PaymentIDs []string `json:"payment_ids"`
}

func (r *ReorderPaymentResourcesRequest) Bind(c echo.Context, u *model.ReorderPaymentResources) error {
	if err := c.Bind(r); err != nil {
		return err
	}
	if len(r.PaymentIDs) == 0 {
		return echo.NewHTTPError(422, "payment_ids is required")
	}
	u.PaymentIDs = r.PaymentIDs
	return nil
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
	u.PaymentTypeId = "1"
	if r.PaymentTypeId != nil {
		u.PaymentTypeId = *r.PaymentTypeId
		u.PaymentDate = r.PaymentDate
		u.ClosingDate = r.ClosingDate
	}

	return nil
}
