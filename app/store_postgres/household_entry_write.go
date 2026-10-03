package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"MoneyHook/MoneyHook-API/model"
	"context"
	"errors"
	"gorm.io/gorm"
	"strings"
	"time"
)

func ptrID(s *string) *uint64 {
	if s == nil {
		return nil
	}
	n := hid(*s)
	return &n
}
func validateFamilyReferences(tx *gorm.DB, f string, in h.EntryInput, payer bool, existing *householdEntryRecord) error {
	var n int64
	if e := tx.Table("category").Where("category_id = ?", in.Transaction.CategoryID).Count(&n).Error; e != nil {
		return e
	}
	if n != 1 {
		return h.Invalid
	}
	for _, ref := range []struct {
		table, column string
		id            *string
		old           *uint64
	}{{"household_payment", "payment_id", in.PaymentID, nil}, {"household_sub_category", "sub_category_id", in.SubCategoryID, nil}} {
		if ref.id == nil {
			continue
		}
		if !h.ID(*ref.id) {
			return h.Invalid
		}
		q := tx.Table(ref.table).Where("household_id = ? AND "+ref.column+" = ?", f, *ref.id)
		allowInactive := existing != nil && ((ref.table == "household_payment" && existing.PaymentID != nil && *existing.PaymentID == hid(*ref.id)) || (ref.table == "household_sub_category" && existing.SubCategoryID != nil && *existing.SubCategoryID == hid(*ref.id)))
		if !allowInactive {
			q = q.Where("active=true")
		}
		if ref.table == "household_sub_category" {
			q = q.Where("category_id = ?", in.Transaction.CategoryID)
		}
		if e := q.Count(&n).Error; e != nil {
			return e
		}
		if n != 1 {
			return h.Invalid
		}
	}
	if payer && in.Payer.Kind == "member" {
		q := tx.Model(&h.Member{}).Where("household_id = ? AND member_id = ?", f, *in.Payer.MemberID)
		if existing == nil || existing.PayerMemberID == nil || *existing.PayerMemberID != hid(*in.Payer.MemberID) {
			q = q.Where("state='active'")
		}
		if e := q.Count(&n).Error; e != nil {
			return e
		}
		if n != 1 {
			return h.Invalid
		}
	}
	return nil
}
func dataFromInput(tx *gorm.DB, in h.EntryInput) (*h.Entry, error) {
	var name string
	e := tx.Table("category").Select("category_name").Where("category_id = ?", in.Transaction.CategoryID).Scan(&name).Error
	if e != nil {
		return nil, e
	}
	t := in.Transaction
	t.Name = strings.TrimSpace(t.Name)
	t.PaymentID = nil
	t.SubCategoryID = ""
	t.SubCategoryName = ""
	return &h.Entry{TransactionInput: t, SignedAmount: t.Amount * int64(t.Sign), CategoryName: name, Excluded: in.Excluded}, nil
}
func (s *HouseholdStore) CreateEntry(ctx context.Context, u, f string, in h.EntryInput, own bool, key string) (*h.Entry, error) {
	if !h.ValidEntry(in, own) {
		return nil, h.Invalid
	}
	var id uint64
	e := householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		if e := lockHouseholdUser(tx, u); e != nil {
			return e
		}
		_, m, e := familyAccess(tx, u, f, true)
		if e != nil {
			return e
		}
		op := "proxy:" + f
		if own {
			op = "own:" + f
		}
		old, e := idempotent(tx, u, op, key, in)
		if e != nil {
			return e
		}
		if old != 0 {
			id = old
			return nil
		}
		references := in
		if own {
			references.PaymentID = nil
			references.SubCategoryID = nil
		}
		if e = validateFamilyReferences(tx, f, references, !own, nil); e != nil {
			return e
		}
		r := householdEntryRecord{HouseholdID: hid(f), Kind: "proxy", PayerMemberID: ptrID(in.Payer.MemberID), PaymentID: ptrID(in.PaymentID), SubCategoryID: ptrID(in.SubCategoryID), State: "active", Version: 1, CreatedBy: m.ID, UpdatedBy: m.ID}
		if own {
			t := in.Transaction
			input := &model.V1TransactionWrite{UserId: u, TransactionDate: t.Date, TransactionTime: t.Time, TransactionName: strings.TrimSpace(t.Name), Amount: t.Amount, Sign: t.Sign, CategoryId: t.CategoryID, SubCategoryId: t.SubCategoryID, SubCategoryName: t.SubCategoryName, FixedFlg: t.Fixed, PaymentId: t.PaymentID}
			created, e := NewTransactionStore(tx).CreateV1Transaction(input)
			if e != nil {
				return e
			}
			source := hid(created.TransactionId)
			r.SourceTransactionID = &source
			r.PayerMemberID = &m.ID
			r.Kind = "shared"
			if e = resolveHouseholdEntryReferences(tx, u, &r); e != nil {
				return e
			}
		}
		if e = tx.Create(&r).Error; e != nil {
			return e
		}
		id = r.ID
		if !own {
			data, e := dataFromInput(tx, in)
			if e != nil {
				return e
			}
			if e = tx.Table("household_entry_data").Create(map[string]any{"entry_id": id, "payload": string(publicEntry(data))}).Error; e != nil {
				return e
			}
		}
		return remember(tx, u, op, key, in, r.ID)
	})
	if e != nil {
		return nil, e
	}
	return s.Entry(ctx, u, f, stringID(id))
}
func (s *HouseholdStore) Share(ctx context.Context, u, f, source string, in h.EntryInput) (*h.Entry, error) {
	var id uint64
	e := householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		if e := lockHouseholdUser(tx, u); e != nil {
			return e
		}
		_, m, e := familyAccess(tx, u, f, true)
		if e != nil {
			return e
		}
		t, e := getPostgresV1Transaction(tx, u, source)
		if e != nil {
			return e
		}
		if e = checkVersion(in.SourceVersion, t.Version); e != nil {
			return e
		}
		var r householdEntryRecord
		e = tx.Where("household_id = ? AND source_transaction_id = ?", f, source).Take(&r).Error
		if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		if r.Kind == "snapshot" {
			return h.Fail("SOURCE_ALREADY_SNAPSHOTTED", "退出時の控えがあるため再共有できません")
		}
		if r.ID != 0 {
			if e = checkVersion(in.ExpectedVersion, r.Version); e != nil {
				return e
			}
		}
		sourceID := hid(source)
		if r.ID == 0 {
			r = householdEntryRecord{HouseholdID: hid(f), Kind: "shared", SourceTransactionID: &sourceID, Version: 1, CreatedBy: m.ID}
		} else {
			r.Version++
		}
		r.PayerMemberID = &m.ID
		if e = resolveHouseholdEntryReferences(tx, u, &r); e != nil {
			return e
		}
		r.State = "active"
		r.UpdatedBy = m.ID
		r.UpdatedAt = time.Now()
		if e = tx.Save(&r).Error; e != nil {
			return e
		}
		id = r.ID
		return nil
	})
	if e != nil {
		return nil, e
	}
	return s.Entry(ctx, u, f, stringID(id))
}
func (s *HouseholdStore) Unshare(ctx context.Context, u, f, source string, v int64) error {
	return householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		if e := lockHouseholdUser(tx, u); e != nil {
			return e
		}
		_, m, e := familyAccess(tx, u, f, true)
		if e != nil {
			return e
		}
		if _, e = getPostgresV1Transaction(tx, u, source); e != nil {
			return h.Forbidden
		}
		var r householdEntryRecord
		if e = tx.Where("household_id = ? AND source_transaction_id = ? AND kind='shared' AND state='active'", f, source).Take(&r).Error; e != nil {
			return h.NotFound
		}
		if e = checkVersion(v, r.Version); e != nil {
			return e
		}
		r.State = "withdrawn"
		r.Version++
		r.UpdatedBy = m.ID
		r.UpdatedAt = time.Now()
		if e = tx.Save(&r).Error; e != nil {
			return e
		}
		return nil
	})
}
