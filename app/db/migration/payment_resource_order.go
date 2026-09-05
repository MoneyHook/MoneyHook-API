package migration

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func backfillPaymentResourceOrder(ctx context.Context, db *gorm.DB) error {
	var rows []paymentResourceSchema
	if err := db.WithContext(ctx).
		Select("payment_id", "user_no", "order_num").
		Order("user_no, payment_id").
		Find(&rows).Error; err != nil {
		return err
	}
	nextByUser := make(map[uint64]int32)
	for _, row := range rows {
		if row.OrderNum > 0 {
			if row.OrderNum >= nextByUser[row.UserNo] {
				nextByUser[row.UserNo] = row.OrderNum + 1
			}
			continue
		}
		nextByUser[row.UserNo]++
		if err := db.WithContext(ctx).Model(&paymentResourceSchema{}).
			Where("payment_id = ?", row.PaymentID).
			Update("order_num", nextByUser[row.UserNo]).Error; err != nil {
			return fmt.Errorf("set order for payment %d: %w", row.PaymentID, err)
		}
	}
	return nil
}
