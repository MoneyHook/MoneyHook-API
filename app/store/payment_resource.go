package store

import (
	"MoneyHook/MoneyHook-API/model"

	"gorm.io/gorm"
)

type PaymentResourceStore struct {
	db *gorm.DB
}

func NewPaymentResourceStore(db *gorm.DB) *PaymentResourceStore {
	return &PaymentResourceStore{db: db}
}

func (pr *PaymentResourceStore) GetPaymentResourceList(userId int) *[]model.PaymentResource {
	var payment_resource_list []model.PaymentResource
	pr.db.Table("payment_resource").
		Where("user_no = ?", userId).
		Find(&payment_resource_list)

	return &payment_resource_list
}
