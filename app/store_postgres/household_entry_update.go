package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

func (s *HouseholdStore) UpdateProxy(ctx context.Context, u, f, id string, in h.EntryInput) (*h.Entry, error) {
	if !h.ValidEntry(in, false) {
		return nil, h.Invalid
	}
	e := s.modifyEntry(ctx, u, f, id, in, false, false)
	if e != nil {
		return nil, e
	}
	return s.Entry(ctx, u, f, id)
}
func (s *HouseholdStore) DeleteProxy(ctx context.Context, u, f, id string, v int64) error {
	return s.modifyEntry(ctx, u, f, id, h.EntryInput{Version: h.Version{ExpectedVersion: v}}, true, false)
}
func (s *HouseholdStore) Correct(ctx context.Context, u, f, id string, in h.EntryInput) (*h.Entry, error) {
	if !h.ValidEntry(in, false) {
		return nil, h.Invalid
	}
	e := s.modifyEntry(ctx, u, f, id, in, false, true)
	if e != nil {
		return nil, e
	}
	return s.Entry(ctx, u, f, id)
}
func (s *HouseholdStore) modifyEntry(ctx context.Context, u, f, id string, in h.EntryInput, remove, correct bool) error {
	return householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		_, m, e := familyAccess(tx, u, f, true)
		if e != nil {
			return e
		}
		r, e := entryRecord(tx, f, id)
		if e != nil {
			return e
		}
		if r.State != "active" {
			return h.NotFound
		}
		if (correct && r.Kind != "snapshot") || (!correct && r.Kind != "proxy") {
			return h.Forbidden
		}
		if e = checkVersion(in.ExpectedVersion, r.Version); e != nil {
			return e
		}
		var after *h.Entry
		if !remove {
			if e = validateFamilyReferences(tx, f, in, true, r); e != nil {
				return e
			}
			after, e = dataFromInput(tx, in)
			if e != nil {
				return e
			}
		}
		r.Version++
		r.UpdatedBy = m.ID
		r.UpdatedAt = time.Now()
		if remove {
			r.State = "deleted"
		} else {
			r.PayerMemberID = ptrID(in.Payer.MemberID)
			r.PaymentID = ptrID(in.PaymentID)
			r.SubCategoryID = ptrID(in.SubCategoryID)
			column := "payload"
			if correct {
				column = "corrected_payload"
				if e = projectEntryReferences(tx, r, after); e != nil {
					return e
				}
			}
			if e = tx.Table("household_entry_data").Where("entry_id = ?", r.ID).Update(column, string(publicEntry(after))).Error; e != nil {
				return e
			}
		}
		return tx.Save(r).Error
	})
}
func captureHouseholdEntries(tx *gorm.DB, f string, m, actor *h.Member, state string) error {
	var rows []householdEntryRecord
	if e := tx.Where("household_id = ? AND payer_member_id = ? AND kind='shared' AND state='active'", f, m.ID).Find(&rows).Error; e != nil {
		return e
	}
	for i := range rows {
		r := &rows[i]
		data, e := entryProjection(tx, r, m)
		if e != nil {
			return e
		}
		now := time.Now()
		data.CapturedAt = &now
		data.Payer.State = state
		sourceVersion := data.SourceVersion
		if e = tx.Table("household_entry_data").Create(map[string]any{"entry_id": r.ID, "payload": string(publicEntry(data)), "captured_at": now, "source_version": sourceVersion}).Error; e != nil {
			return e
		}
		r.Kind = "snapshot"
		r.UpdatedBy = actor.ID
		r.Version++
		r.UpdatedAt = now
		if e = tx.Save(r).Error; e != nil {
			return e
		}
	}
	return nil
}

func (s *HouseholdStore) ShareStatus(ctx context.Context, u, f, source string) (*h.ShareStatus, error) {
	tx := s.db.WithContext(ctx)
	if _, _, e := familyAccess(tx, u, f, false); e != nil {
		return nil, e
	}
	if _, e := getPostgresV1Transaction(tx, u, source); e != nil {
		return nil, h.NotFound
	}
	var r householdEntryRecord
	e := tx.Where("household_id = ? AND source_transaction_id = ?", f, source).Take(&r).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return &h.ShareStatus{State: "none"}, nil
	}
	if e != nil {
		return nil, e
	}
	return &h.ShareStatus{EntryID: stringID(r.ID), Version: r.Version, State: r.State, Kind: r.Kind}, nil
}
