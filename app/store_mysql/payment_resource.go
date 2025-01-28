package store_mysql

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
		Order("payment_id").
		Find(&payment_resource_list)

	for i, item := range payment_resource_list {
		if item.ClosingDate == 0 {
			payment_resource_list[i].ClosingDate = 31
		}
	}

	return &payment_resource_list
}

func (pr *PaymentResourceStore) AddPaymentResource(addPayment *model.AddPaymentResource) error {
	return pr.db.Table("payment_resource").Create(&addPayment).Error
}

func (pr *PaymentResourceStore) EditPaymentResource(editPayment *model.EditPaymentResource) error {
	return pr.db.Table("payment_resource").
		Where("payment_id = ?", editPayment.PaymentId).
		Where("user_no =?", editPayment.UserNo).
		Update("payment_name", editPayment.PaymentName).
		Update("payment_type_id", editPayment.PaymentTypeId).
		Update("payment_date", editPayment.PaymentDate).
		Update("closing_date", editPayment.ClosingDate).
		Error
}

func (pr *PaymentResourceStore) DeletePaymentResource(deletePayment *model.DeletePaymentResource) error {
	return pr.db.Table("payment_resource").
		Where("payment_id = ?", deletePayment.PaymentId).
		Where("user_no = ?", deletePayment.UserNo).
		Delete(&model.DeletePaymentResource{}).
		Error
}

func (pr *PaymentResourceStore) GetPaymentTypeList() *[]model.PaymentType {
	var payment_type_list []model.PaymentType
	pr.db.Table("payment_type").
		Order("order_num").
		Find(&payment_type_list)

	return &payment_type_list
}
