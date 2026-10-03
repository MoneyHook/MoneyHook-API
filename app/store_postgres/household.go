package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"time"
)

type HouseholdStore struct {
	db     *gorm.DB
	secret []byte
}

func NewHouseholdStore(db *gorm.DB, secret string) *HouseholdStore {
	return &HouseholdStore{db, []byte(secret)}
}
func hid(s string) uint64 { v, _ := strconv.ParseUint(s, 10, 64); return v }
func lockHouseholdUser(tx *gorm.DB, u string) error {
	var id uint64
	return tx.Table("users").Select("user_no").Where("user_no = ?", u).Clauses(clause.Locking{Strength: "UPDATE"}).Take(&id).Error
}
func familyAccess(tx *gorm.DB, u, f string, write bool) (*h.Family, *h.Member, error) {
	var fam h.Family
	q := tx.Where("household_id = ?", f)
	if write {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := q.Take(&fam).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, h.NotFound
		}
		return nil, nil, err
	}
	var m h.Member
	if err := tx.Where("household_id = ? AND user_no = ? AND state IN ('active','archived')", f, u).Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, h.NotFound
		}
		return nil, nil, err
	}
	if write && (fam.State != "active" || m.State != "active") {
		return nil, nil, h.Forbidden
	}
	fam.Role = m.Role
	fam.MemberID = m.ID
	return &fam, &m, nil
}
func requireAdmin(m *h.Member) error {
	if m.Role != "admin" {
		return h.Forbidden
	}
	return nil
}
func checkVersion(expected, actual int64) error {
	if expected < 1 {
		return h.Invalid
	}
	if expected != actual {
		return h.Conflict
	}
	return nil
}
func bumpFamily(tx *gorm.DB, f string) error {
	return tx.Model(&h.Family{}).Where("household_id = ?", f).Update("version", gorm.Expr("version + 1")).Error
}
func digest(v any) string {
	b, _ := json.Marshal(v)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
func idempotent(tx *gorm.DB, u, op, key string, body any) (uint64, error) {
	if len(key) < 8 || len(key) > 128 {
		return 0, h.Fail("VALIDATION_ERROR", "Idempotency-Keyが必要です")
	}
	var r struct {
		ResourceID    uint64
		RequestDigest string
	}
	err := tx.Table("api_idempotency").Where("user_no = ? AND operation = ? AND key = ?", u, op, key).Take(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if r.RequestDigest != digest(body) {
		return 0, h.Fail("IDEMPOTENCY_CONFLICT", "同じキーで異なる内容は保存できません")
	}
	return r.ResourceID, nil
}
func remember(tx *gorm.DB, u, op, key string, body any, id uint64) error {
	return tx.Table("api_idempotency").Create(map[string]any{"user_no": u, "operation": op, "key": key, "request_digest": digest(body), "resource_id": id}).Error
}
func (s *HouseholdStore) List(ctx context.Context, u string) ([]h.Family, error) {
	var ms []h.Member
	if err := s.db.WithContext(ctx).Where("user_no = ? AND state IN ('active','archived')", u).Find(&ms).Error; err != nil {
		return nil, err
	}
	out := []h.Family{}
	for _, m := range ms {
		f, err := s.Get(ctx, u, stringID(m.HouseholdID))
		if err != nil {
			return nil, err
		}
		out = append(out, *f)
	}
	return out, nil
}
func (s *HouseholdStore) Get(ctx context.Context, u, f string) (*h.Family, error) {
	fam, _, err := familyAccess(s.db.WithContext(ctx), u, f, false)
	return fam, err
}
func (s *HouseholdStore) Create(ctx context.Context, u string, in h.FamilyInput, key string) (*h.Family, error) {
	if !h.Name(in.Name, 64) || !h.Name(in.DisplayName, 32) {
		return nil, h.Invalid
	}
	var out *h.Family
	err := householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		if err := lockHouseholdUser(tx, u); err != nil {
			return err
		}
		id, err := idempotent(tx, u, "family-create", key, in)
		if err != nil {
			return err
		}
		if id != 0 {
			out, _, err = familyAccess(tx, u, stringID(id), true)
			return err
		}
		var n int64
		if err := tx.Model(&h.Member{}).Where("user_no = ? AND state='active'", u).Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return h.Fail("ALREADY_IN_HOUSEHOLD", "参加できる家族は1つです")
		}
		out = &h.Family{Name: strings.TrimSpace(in.Name), State: "active", Version: 1}
		if err := tx.Create(out).Error; err != nil {
			return err
		}
		slot := 1
		m := h.Member{HouseholdID: out.ID, UserNo: hid(u), Role: "admin", State: "active", SlotNo: &slot, DisplayName: strings.TrimSpace(in.DisplayName), JoinedAt: time.Now()}
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		out.Role = m.Role
		out.MemberID = m.ID
		return remember(tx, u, "family-create", key, in, out.ID)
	})
	return out, err
}
func (s *HouseholdStore) Rename(ctx context.Context, u, f string, in h.FamilyInput) (*h.Family, error) {
	if !h.Name(in.Name, 64) {
		return nil, h.Invalid
	}
	err := householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		fam, m, err := familyAccess(tx, u, f, true)
		if err != nil {
			return err
		}
		if err = requireAdmin(m); err != nil {
			return err
		}
		if err = checkVersion(in.ExpectedVersion, fam.Version); err != nil {
			return err
		}
		return tx.Model(fam).Updates(map[string]any{"name": strings.TrimSpace(in.Name), "version": fam.Version + 1}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, u, f)
}
func (s *HouseholdStore) Members(ctx context.Context, u, f string) ([]h.Member, error) {
	tx := s.db.WithContext(ctx)
	if _, _, err := familyAccess(tx, u, f, false); err != nil {
		return nil, err
	}
	out := []h.Member{}
	err := tx.Where("household_id = ?", f).Order("member_id").Find(&out).Error
	return out, err
}
func (s *HouseholdStore) RenameMember(ctx context.Context, u, f, name string) error {
	if !h.Name(name, 32) {
		return h.Invalid
	}
	return householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		_, m, e := familyAccess(tx, u, f, true)
		if e != nil {
			return e
		}
		return tx.Model(m).Update("display_name", strings.TrimSpace(name)).Error
	})
}
func (s *HouseholdStore) Transfer(ctx context.Context, u, f string, in h.MemberInput) error {
	return householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		fam, m, e := familyAccess(tx, u, f, true)
		if e != nil {
			return e
		}
		if e = requireAdmin(m); e != nil {
			return e
		}
		if e = checkVersion(in.ExpectedVersion, fam.Version); e != nil {
			return e
		}
		var next h.Member
		if e = tx.Where("household_id = ? AND member_id = ? AND state='active' AND role='member'", f, in.MemberID).Take(&next).Error; e != nil {
			return h.Invalid
		}
		if e = tx.Model(m).Update("role", "member").Error; e != nil {
			return e
		}
		if e = tx.Model(&next).Update("role", "admin").Error; e != nil {
			return e
		}
		if e = tx.Model(&h.Invitation{}).Where("household_id = ? AND consumed_at IS NULL AND revoked_at IS NULL", f).Update("revoked_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; e != nil {
			return e
		}
		return bumpFamily(tx, f)
	})
}
func (s *HouseholdStore) Leave(ctx context.Context, u, f, target string, version int64, archive bool) error {
	tx := s.db.WithContext(ctx)
	var candidate h.Member
	q := tx.Where("household_id = ?", f)
	if target == "" {
		q = q.Where("user_no = ?", u)
	} else {
		q = q.Where("member_id = ?", target)
	}
	if e := q.Take(&candidate).Error; e != nil {
		return h.NotFound
	}
	return householdTransaction(tx, func(tx *gorm.DB) error {
		if e := lockHouseholdUser(tx, stringID(candidate.UserNo)); e != nil {
			return e
		}
		fam, m, e := familyAccess(tx, u, f, true)
		if e != nil {
			return e
		}
		if e = checkVersion(version, fam.Version); e != nil {
			return e
		}
		var subject h.Member
		if e = tx.Where("member_id = ? AND state='active'", candidate.ID).Take(&subject).Error; e != nil {
			return h.NotFound
		}
		if subject.UserNo != hid(u) && m.Role != "admin" {
			return h.Forbidden
		}
		state := "left"
		if archive {
			if subject.ID != m.ID || m.Role != "admin" {
				return h.Forbidden
			}
			var n int64
			if e = tx.Model(&h.Member{}).Where("household_id = ? AND state='active'", f).Count(&n).Error; e != nil {
				return e
			}
			if n != 1 {
				return h.Fail("MEMBERS_REMAIN", "他のメンバーが参加しています")
			}
			state = "archived"
		} else if subject.Role == "admin" {
			return h.Fail("ADMIN_TRANSFER_REQUIRED", "退出前に管理者を交代してください")
		}
		if e = captureHouseholdEntries(tx, f, &subject, m, state); e != nil {
			return e
		}
		if e = tx.Model(&subject).Updates(map[string]any{"state": state, "slot_no": nil, "left_at": gorm.Expr("CURRENT_TIMESTAMP")}).Error; e != nil {
			return e
		}
		if e = tx.Table("users").Where("user_no = ?", subject.UserNo).Update("default_transaction_scope", "personal").Error; e != nil {
			return e
		}
		if archive {
			if e = tx.Model(fam).Updates(map[string]any{"state": "archived", "version": fam.Version + 1, "archived_at": gorm.Expr("CURRENT_TIMESTAMP")}).Error; e != nil {
				return e
			}
			if e = tx.Model(&h.Invitation{}).Where("household_id = ? AND consumed_at IS NULL", f).Update("revoked_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; e != nil {
				return e
			}
		} else {
			if e = bumpFamily(tx, f); e != nil {
				return e
			}
		}
		return nil
	})
}
