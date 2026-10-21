package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// BulkShare sets the requested state using current records under the same locks
// as single-item sharing and personal transaction edits.
func (s *HouseholdStore) BulkShare(ctx context.Context, u, f string, in h.BulkSharingInput, share bool) (*h.BulkSharingResponse, error) {
	if !h.ValidBulkSharing(in) {
		return nil, h.Invalid
	}
	var out h.BulkSharingResponse
	err := householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		out.Results = make([]h.BulkSharingResult, 0, len(in.TransactionIDs))
		if err := lockHouseholdUser(tx, u); err != nil {
			return err
		}
		_, member, err := familyAccess(tx, u, f, true)
		if err != nil {
			return err
		}
		var originals []struct {
			ID         uint64                      `gorm:"column:transaction_id"`
			References householdPersonalReferences `gorm:"embedded"`
		}
		if err = tx.Table("transaction").Select("transaction_id,category_id,sub_category_id,payment_id").Where("user_no = ? AND deleted_at IS NULL AND transaction_id IN ?", u, in.TransactionIDs).Find(&originals).Error; err != nil {
			return err
		}
		owned := make(map[uint64]householdPersonalReferences, len(originals))
		for _, row := range originals {
			owned[row.ID] = row.References
		}
		var entries []householdEntryRecord
		if err = tx.Where("household_id = ? AND source_transaction_id IN ?", f, in.TransactionIDs).Find(&entries).Error; err != nil {
			return err
		}
		existing := make(map[uint64]householdEntryRecord, len(entries))
		for _, row := range entries {
			existing[*row.SourceTransactionID] = row
		}
		seen := make(map[uint64]bool)
		for _, id := range in.TransactionIDs {
			source := hid(id)
			if seen[source] {
				continue
			}
			seen[source] = true
			result := h.BulkSharingResult{TransactionID: id}
			row := existing[source]
			refs, isOwned := owned[source]
			switch {
			case !isOwned:
				result = bulkSharingFailure(id, h.NotFound)
			case row.Kind == "snapshot":
				result = bulkSharingFailure(id, h.Fail("SOURCE_ALREADY_SNAPSHOTTED", "退出時の控えは共有・共有解除できません"))
			case share && row.State == "active" && row.Kind == "shared":
				result.Outcome = "already-shared"
			case !share && row.State != "active":
				result.Outcome = "already-unshared"
			case row.ID != 0 && row.Kind != "shared":
				result = bulkSharingFailure(id, h.Forbidden)
			default:
				// A failed reference creation or write must not leak into the next item.
				if err := tx.SavePoint("bulk_item").Error; err != nil {
					return err
				}
				err := writeBulkSharingItem(tx, u, f, source, member, row, refs, share)
				if err != nil {
					var pg *pgconn.PgError
					if errors.As(err, &pg) && (pg.Code == "40001" || pg.Code == "40P01") {
						return err
					}
					if rollbackErr := tx.RollbackTo("bulk_item").Error; rollbackErr != nil {
						return rollbackErr
					}
					result = bulkSharingFailure(id, err)
				} else if share {
					result.Outcome = "shared"
				} else {
					result.Outcome = "unshared"
				}
				if err := tx.Exec("RELEASE SAVEPOINT bulk_item").Error; err != nil {
					return err
				}
			}
			out.Results = append(out.Results, result)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func writeBulkSharingItem(tx *gorm.DB, u, f string, source uint64, member *h.Member, row householdEntryRecord, refs householdPersonalReferences, share bool) error {
	if share {
		if row.ID == 0 {
			row = householdEntryRecord{HouseholdID: hid(f), Kind: "shared", SourceTransactionID: &source, CreatedBy: member.ID}
		}
		row.PayerMemberID = &member.ID
		var err error
		row.SubCategoryID, err = resolveHouseholdSubcategory(tx, row.HouseholdID, refs)
		if err != nil {
			return err
		}
		row.PaymentID, err = resolveHouseholdPayment(tx, row.HouseholdID, u, refs.PaymentID)
		if err != nil {
			return err
		}
		row.State = "active"
	} else {
		row.State = "withdrawn"
	}
	row.Version++
	row.UpdatedBy = member.ID
	row.UpdatedAt = time.Now()
	return tx.Save(&row).Error
}

func bulkSharingFailure(id string, err error) h.BulkSharingResult {
	result := h.BulkSharingResult{TransactionID: id, Outcome: "failed", Code: "INTERNAL_ERROR", Message: "処理に失敗しました"}
	var domain *h.Error
	if errors.As(err, &domain) {
		result.Code, result.Message = domain.Code, domain.Message
	}
	return result
}
