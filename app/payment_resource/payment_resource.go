package sub_category

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetPaymentResourceList(userId int) *[]model.PaymentResource
	AddPaymentResource(*model.AddPaymentResource) error
	DeletePaymentResource(*model.DeletePaymentResource) error
}
