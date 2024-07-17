package response

import "MoneyHook/MoneyHook-API/model"

/*
支払い方法
*/
type paymentResourceListResponse struct {
	PaymentResourceList []paymentResourceResponse `json:"payment_list"`
}
type paymentResourceResponse struct {
	PaymentId     int    `json:"payment_id"`
	PaymentName   string `json:"payment_name"`
	PaymentTypeId int    `json:"payment_type_id"`
	PaymentDate   *int   `json:"payment_date"`
	ClosingDate   *int   `json:"closing_date"`
}

func GetPaymentResourceListResponse(data *[]model.PaymentResource) *paymentResourceListResponse {
	prl := &paymentResourceListResponse{PaymentResourceList: []paymentResourceResponse{}}

	for _, payment_resource := range *data {
		var paymentDate *int
		if payment_resource.PaymentDate != 0 {
			paymentDate = &payment_resource.PaymentDate
		}
		var closingDate *int
		if payment_resource.ClosingDate != 0 {
			closingDate = &payment_resource.ClosingDate
		}

		scr := &paymentResourceResponse{
			PaymentId:     payment_resource.PaymentId,
			PaymentName:   payment_resource.PaymentName,
			PaymentTypeId: payment_resource.PaymentTypeId,
			PaymentDate:   paymentDate,
			ClosingDate:   closingDate,
		}
		prl.PaymentResourceList = append(prl.PaymentResourceList, *scr)
	}

	return prl
}

type paymentTypeListResponse struct {
	PaymentTypeList []paymentTypeResponse `json:"payment_type_list"`
}
type paymentTypeResponse struct {
	PaymentTypeId     int    `json:"payment_type_id"`
	PaymentTypeName   string `json:"payment_type_name"`
	IsPaymentDueLater bool   `json:"is_payment_due_later"`
}

func GetPaymentTypeListResponse(data *[]model.PaymentType) *paymentTypeListResponse {
	ptl := &paymentTypeListResponse{PaymentTypeList: []paymentTypeResponse{}}

	for _, payment_type := range *data {
		ptr := &paymentTypeResponse{
			PaymentTypeId:     payment_type.PaymentTypeId,
			PaymentTypeName:   payment_type.PaymentTypeName,
			IsPaymentDueLater: payment_type.IsPaymentDueLater}
		ptl.PaymentTypeList = append(ptl.PaymentTypeList, *ptr)
	}

	return ptl
}
