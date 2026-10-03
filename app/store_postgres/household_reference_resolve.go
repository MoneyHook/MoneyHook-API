package store_postgres

import (
	"errors"

	"gorm.io/gorm"
)

// Personal IDs are only used inside the save transaction, never exposed by a
// family response. Callers hold the family row lock via familyAccess(write=true).
type householdPersonalReferences struct {
	CategoryID    string  `gorm:"column:category_id"`
	SubCategoryID string  `gorm:"column:sub_category_id"`
	PaymentID     *string `gorm:"column:payment_id"`
}

func sourceHouseholdReferences(tx *gorm.DB, source uint64) (householdPersonalReferences, error) {
	var refs householdPersonalReferences
	err := tx.Table("transaction").Select("category_id,sub_category_id,payment_id").Where("transaction_id = ? AND deleted_at IS NULL", source).Take(&refs).Error
	return refs, err
}

func resolveHouseholdSubcategory(tx *gorm.DB, family uint64, refs householdPersonalReferences) (*uint64, error) {
	var name string
	if err := tx.Table("sub_category").Select("sub_category_name").Where("sub_category_id = ? AND category_id = ?", refs.SubCategoryID, refs.CategoryID).Take(&name).Error; err != nil {
		return nil, err
	}
	return resolveHouseholdReference(tx, "household_sub_category", "sub_category_id", map[string]any{
		"household_id": family, "category_id": refs.CategoryID, "name": name,
	})
}

func resolveHouseholdPayment(tx *gorm.DB, family uint64, user string, payment *string) (*uint64, error) {
	if payment == nil {
		return nil, nil
	}
	var p struct {
		Name          string
		PaymentTypeID string
		PaymentDate   *int
		ClosingDate   *int
	}
	// Match the effective dates used by the personal payment API.
	if err := tx.Table("payment_resource").Select("payment_name AS name,payment_type_id,NULLIF(payment_date,0) AS payment_date,COALESCE(NULLIF(closing_date,0),31) AS closing_date").Where("payment_id = ? AND user_no = ?", *payment, user).Take(&p).Error; err != nil {
		return nil, err
	}
	return resolveHouseholdReference(tx, "household_payment", "payment_id", map[string]any{
		"household_id": family, "name": p.Name, "payment_type_id": p.PaymentTypeID,
		"payment_date": p.PaymentDate, "closing_date": p.ClosingDate,
	})
}

func resolveHouseholdReference(tx *gorm.DB, table, column string, values map[string]any) (*uint64, error) {
	values["active"] = true
	var row struct{ ID uint64 }
	// Existing duplicate definitions remain intact; pick the oldest active match.
	err := tx.Table(table).Select(column + " AS id").Where(values).Order(column).Take(&row).Error
	if err == nil {
		return &row.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// The family lock also serializes manual reference creation and activation.
	if err := tx.Table(table).Create(values).Error; err != nil {
		return nil, err
	}
	if err := tx.Table(table).Select(column + " AS id").Where(values).Order(column).Take(&row).Error; err != nil {
		return nil, err
	}
	return &row.ID, nil
}

func resolveHouseholdEntryReferences(tx *gorm.DB, user string, r *householdEntryRecord) error {
	refs, err := sourceHouseholdReferences(tx, *r.SourceTransactionID)
	if err != nil {
		return err
	}
	r.SubCategoryID, err = resolveHouseholdSubcategory(tx, r.HouseholdID, refs)
	if err != nil {
		return err
	}
	r.PaymentID, err = resolveHouseholdPayment(tx, r.HouseholdID, user, refs.PaymentID)
	return err
}

func samePersonalPayment(a, b *string) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}
