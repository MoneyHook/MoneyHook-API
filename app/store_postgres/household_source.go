package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"gorm.io/gorm"
	"time"
)

type householdSourceChange struct {
	Record     householdEntryRecord
	Member     h.Member
	References householdPersonalReferences
}

func beginHouseholdSourceChange(tx *gorm.DB, u, id string) ([]householdSourceChange, error) {
	if e := lockHouseholdUser(tx, u); e != nil {
		return nil, e
	}
	var rows []householdEntryRecord
	if e := tx.Where("source_transaction_id = ? AND kind='shared' AND state='active'", id).Find(&rows).Error; e != nil {
		return nil, e
	}
	changes := []householdSourceChange{}
	for _, r := range rows {
		_, m, e := familyAccess(tx, u, stringID(r.HouseholdID), true)
		if e != nil {
			return nil, e
		}
		refs, e := sourceHouseholdReferences(tx, *r.SourceTransactionID)
		if e != nil {
			return nil, e
		}
		changes = append(changes, householdSourceChange{r, *m, refs})
	}
	return changes, nil
}
func finishHouseholdSourceChange(tx *gorm.DB, changes []householdSourceChange, deleted bool) error {
	for _, c := range changes {
		r := c.Record
		r.Version++
		r.UpdatedBy = c.Member.ID
		r.UpdatedAt = time.Now()
		if deleted {
			r.State = "withdrawn"
		} else {
			refs, e := sourceHouseholdReferences(tx, *r.SourceTransactionID)
			if e != nil {
				return e
			}
			if refs.CategoryID != c.References.CategoryID || refs.SubCategoryID != c.References.SubCategoryID {
				r.SubCategoryID, e = resolveHouseholdSubcategory(tx, r.HouseholdID, refs)
				if e != nil {
					return e
				}
			}
			if !samePersonalPayment(refs.PaymentID, c.References.PaymentID) {
				r.PaymentID, e = resolveHouseholdPayment(tx, r.HouseholdID, stringID(c.Member.UserNo), refs.PaymentID)
				if e != nil {
					return e
				}
			}
		}
		if e := tx.Save(&r).Error; e != nil {
			return e
		}
	}
	return nil
}
