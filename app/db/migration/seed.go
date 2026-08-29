package migration

import (
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func seedMasterData(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := ensureSystemUser(tx); err != nil {
			return err
		}

		categoryResult := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "category_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"category_name",
				"order_num",
			}),
		}).Create(&masterCategories)
		if categoryResult.Error != nil {
			return categoryResult.Error
		}

		paymentTypeResult := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "payment_type_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"payment_type_name",
				"is_payment_due_later",
				"order_num",
			}),
		}).Create(&masterPaymentTypes)
		if paymentTypeResult.Error != nil {
			return paymentTypeResult.Error
		}

		subCategories := make([]subCategorySchema, 0, len(masterSubCategories))
		for _, master := range masterSubCategories {
			subCategories = append(subCategories, subCategorySchema{
				UserNo:          systemUserNo,
				CategoryID:      master.CategoryID,
				SubCategoryName: master.Name,
			})
		}
		subCategoryResult := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&subCategories)
		if subCategoryResult.Error != nil {
			return subCategoryResult.Error
		}

		log.Printf(
			"event=schema_migration_master_data categories=%d payment_types=%d sub_categories_inserted=%d",
			len(masterCategories),
			len(masterPaymentTypes),
			subCategoryResult.RowsAffected,
		)
		return nil
	})
}

func ensureSystemUser(db *gorm.DB) error {
	var byNumber userSchema
	result := db.Where("user_no = ?", systemUserNo).Limit(1).Find(&byNumber)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		if byNumber.UserID != systemUserID {
			return fmt.Errorf("user_no=%d belongs to user_id=%q instead of the system user", systemUserNo, byNumber.UserID)
		}
		return nil
	}

	var count int64
	if err := db.Model(&userSchema{}).Where("user_id = ?", systemUserID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("system user_id=%q already exists with a user_no other than %d", systemUserID, systemUserNo)
	}
	user := userSchema{UserNo: systemUserNo, UserID: systemUserID}
	if err := db.Create(&user).Error; err != nil {
		return err
	}
	log.Printf("event=schema_migration_change action=create_system_user user_no=%d", systemUserNo)
	return nil
}
