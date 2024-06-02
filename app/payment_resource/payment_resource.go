package sub_category

import "MoneyHook/MoneyHook-API/model"

type Store interface {
	GetPaymentResourceList(userId int) *[]model.PaymentResource
}
