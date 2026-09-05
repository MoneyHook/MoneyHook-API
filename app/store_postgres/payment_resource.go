package store_postgres

import (
	"MoneyHook/MoneyHook-API/model"
	"fmt"

	"gorm.io/gorm"
)

type PaymentResourceStore struct {
	db *gorm.DB
}

func NewPaymentResourceStore(db *gorm.DB) *PaymentResourceStore {
	return &PaymentResourceStore{db: db}
}

func (pr *PaymentResourceStore) GetPaymentResourceList(userId string) *[]model.PaymentResource {
	var payment_resource_list []model.PaymentResource
	pr.db.Table("payment_resource").
		Where("user_no = ?", userId).
		Order("order_num, payment_id").
		Find(&payment_resource_list)

	for i, item := range payment_resource_list {
		if item.ClosingDate == 0 {
			payment_resource_list[i].ClosingDate = 31
		}
	}

	return &payment_resource_list
}

func (pr *PaymentResourceStore) AddPaymentResource(addPayment *model.AddPaymentResource) error {
	return pr.db.Transaction(func(tx *gorm.DB) error {
		var maximumOrder int
		if err := tx.Table("payment_resource").
			Select("COALESCE(MAX(order_num), 0)").
			Where("user_no = ?", addPayment.UserNo).
			Scan(&maximumOrder).Error; err != nil {
			return err
		}
		addPayment.OrderNum = maximumOrder + 1
		return tx.Table("payment_resource").Create(addPayment).Error
	})
}

func (pr *PaymentResourceStore) ReorderPaymentResources(reorder *model.ReorderPaymentResources) error {
	if len(reorder.PaymentIDs) == 0 {
		return fmt.Errorf("payment order must not be empty")
	}
	seen := make(map[string]struct{}, len(reorder.PaymentIDs))
	for _, paymentID := range reorder.PaymentIDs {
		if _, exists := seen[paymentID]; exists {
			return fmt.Errorf("payment order contains duplicates")
		}
		seen[paymentID] = struct{}{}
	}
	return pr.db.Transaction(func(tx *gorm.DB) error {
		var owned []model.PaymentResource
		if err := tx.Table("payment_resource").
			Select("payment_id").
			Where("user_no = ?", reorder.UserNo).
			Find(&owned).Error; err != nil {
			return err
		}
		if len(owned) != len(reorder.PaymentIDs) {
			return fmt.Errorf("payment order does not match owned payment resources")
		}
		for _, payment := range owned {
			if _, exists := seen[payment.PaymentId]; !exists {
				return fmt.Errorf("payment order contains an unowned payment resource")
			}
		}
		for index, paymentID := range reorder.PaymentIDs {
			if err := tx.Table("payment_resource").
				Where("payment_id = ?", paymentID).
				Where("user_no = ?", reorder.UserNo).
				Update("order_num", index+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
