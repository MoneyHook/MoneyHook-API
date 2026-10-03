package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
)

var inviteUnavailable = h.Fail("INVITATION_UNAVAILABLE", "招待が無効です。期限またはコードを確認してください")

func (s *HouseholdStore) inviteDigest(kind, value string) string {
	m := hmac.New(sha256.New, s.secret)
	m.Write([]byte(kind + ":" + value))
	return hex.EncodeToString(m.Sum(nil))
}
func (s *HouseholdStore) invitationReady() error {
	if len(s.secret) < 32 {
		return h.Fail("INVITATION_NOT_CONFIGURED", "招待機能の設定が完了していません")
	}
	return nil
}
func (s *HouseholdStore) Invitations(ctx context.Context, u, f string) ([]h.Invitation, error) {
	tx := s.db.WithContext(ctx)
	_, m, e := familyAccess(tx, u, f, false)
	if e != nil {
		return nil, e
	}
	if e = requireAdmin(m); e != nil {
		return nil, e
	}
	rows := []h.Invitation{}
	e = tx.Where("household_id = ?", f).Order("invitation_id DESC").Limit(100).Find(&rows).Error
	for i := range rows {
		rows[i].State = "active"
		switch {
		case rows[i].ConsumedAt != nil:
			rows[i].State = "used"
		case rows[i].RevokedAt != nil:
			rows[i].State = "revoked"
		case !time.Now().Before(rows[i].ExpiresAt):
			rows[i].State = "expired"
		}
	}
	return rows, e
}
func (s *HouseholdStore) Issue(ctx context.Context, u, f, replace, key string) (*h.InviteSecret, error) {
	if e := s.invitationReady(); e != nil {
		return nil, e
	}
	var out *h.InviteSecret
	e := householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		_, m, e := familyAccess(tx, u, f, true)
		if e != nil {
			return e
		}
		if e = requireAdmin(m); e != nil {
			return e
		}
		op := "invite:" + f
		old, e := idempotent(tx, u, op, key, replace)
		if e != nil {
			return e
		}
		if old != 0 {
			return h.Fail("INVITATION_ALREADY_ISSUED", "発行済みです。招待一覧から再発行してください")
		}
		if replace != "" {
			r := tx.Model(&h.Invitation{}).Where("household_id = ? AND invitation_id = ? AND consumed_at IS NULL", f, replace).Update("revoked_at", gorm.Expr("CURRENT_TIMESTAMP"))
			if r.Error != nil {
				return r.Error
			}
			if r.RowsAffected == 0 {
				return inviteUnavailable
			}
		}
		var members, invites int64
		if e = tx.Model(&h.Member{}).Where("household_id = ? AND state='active'", f).Count(&members).Error; e != nil {
			return e
		}
		if e = tx.Model(&h.Invitation{}).Where("household_id = ? AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP", f).Count(&invites).Error; e != nil {
			return e
		}
		if members+invites >= 3 {
			return h.Fail("INVITATION_LIMIT_REACHED", "人数枠がありません。不要な招待を取り消してください")
		}
		var random [42]byte
		if _, e = rand.Read(random[:]); e != nil {
			return e
		}
		alphabet := "0123456789ABCDEFGHJKMNPQRSTVWXYZ" // 32 symbols, excludes I/L/O/U.
		code := make([]byte, 10)
		for i := range code {
			code[i] = alphabet[random[i]&31]
		}
		token := base64.RawURLEncoding.EncodeToString(random[10:])
		var expires time.Time
		if e = tx.Raw("SELECT CURRENT_TIMESTAMP + INTERVAL '24 hours'").Scan(&expires).Error; e != nil {
			return e
		}
		inv := h.Invitation{HouseholdID: hid(f), IssuedBy: m.ID, CodeDigest: s.inviteDigest("code", string(code)), TokenDigest: s.inviteDigest("token", token), ExpiresAt: expires}
		if e = tx.Create(&inv).Error; e != nil {
			return e
		}
		inv.State = "active"
		out = &h.InviteSecret{Invitation: inv, Code: string(code[:5]) + "-" + string(code[5:]), Token: token}
		return remember(tx, u, op, key, replace, inv.ID)
	})
	return out, e
}
func (s *HouseholdStore) Revoke(ctx context.Context, u, f, id string) error {
	return householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		_, m, e := familyAccess(tx, u, f, true)
		if e != nil {
			return e
		}
		if e = requireAdmin(m); e != nil {
			return e
		}
		r := tx.Model(&h.Invitation{}).Where("household_id = ? AND invitation_id = ? AND consumed_at IS NULL", f, id).Update("revoked_at", gorm.Expr("CURRENT_TIMESTAMP"))
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected == 0 {
			return inviteUnavailable
		}
		return nil
	})
}

// Persist limits separately from acceptance: failed guesses must count too.
func (s *HouseholdStore) limitInvite(ctx context.Context, u, ip string) error {
	for _, v := range []string{"user:" + u, "ip:" + ip} {
		var n int
		e := s.db.WithContext(ctx).Raw(`INSERT INTO household_invitation_attempt(bucket,attempts,expires_at) VALUES (?,1,CURRENT_TIMESTAMP + INTERVAL '15 minutes') ON CONFLICT(bucket) DO UPDATE SET attempts=CASE WHEN household_invitation_attempt.expires_at <= CURRENT_TIMESTAMP THEN 1 ELSE household_invitation_attempt.attempts+1 END, expires_at=CASE WHEN household_invitation_attempt.expires_at <= CURRENT_TIMESTAMP THEN CURRENT_TIMESTAMP + INTERVAL '15 minutes' ELSE household_invitation_attempt.expires_at END RETURNING attempts`, digest(v)).Scan(&n).Error
		if e != nil {
			return e
		}
		if n > 20 {
			return h.Fail("RATE_LIMITED", "試行回数を超えました。15分後に再度お試しください")
		}
	}
	return nil
}
func (s *HouseholdStore) findInvitation(tx *gorm.DB, c h.Credential) (*h.Invitation, error) {
	if !h.ValidCredential(c) {
		return nil, inviteUnavailable
	}
	var inv h.Invitation
	q := tx.Where("consumed_at IS NULL AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP")
	if c.Code != "" {
		q = q.Where("code_digest = ?", s.inviteDigest("code", h.NormalizeCode(c.Code)))
	} else {
		q = q.Where("token_digest = ?", s.inviteDigest("token", c.Token))
	}
	if e := q.Take(&inv).Error; e != nil {
		if errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, inviteUnavailable
		}
		return nil, e
	}
	return &inv, nil
}
func (s *HouseholdStore) Preview(ctx context.Context, u, ip string, c h.Credential) (*h.Preview, error) {
	if e := s.invitationReady(); e != nil {
		return nil, e
	}
	if e := s.limitInvite(ctx, u, ip); e != nil {
		return nil, e
	}
	tx := s.db.WithContext(ctx)
	inv, e := s.findInvitation(tx, c)
	if e != nil {
		return nil, e
	}
	var f h.Family
	var m h.Member
	if e = tx.Where("household_id = ? AND state='active'", inv.HouseholdID).Take(&f).Error; e != nil {
		return nil, inviteUnavailable
	}
	if e = tx.Where("member_id = ? AND role='admin' AND state='active'", inv.IssuedBy).Take(&m).Error; e != nil {
		return nil, inviteUnavailable
	}
	return &h.Preview{Name: f.Name, Inviter: m.DisplayName, ExpiresAt: inv.ExpiresAt}, nil
}
func (s *HouseholdStore) Accept(ctx context.Context, u, ip string, c h.Credential, key string) (*h.Family, error) {
	if e := s.invitationReady(); e != nil {
		return nil, e
	}
	if e := s.limitInvite(ctx, u, ip); e != nil {
		return nil, e
	}
	var out *h.Family
	e := householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		if e := lockHouseholdUser(tx, u); e != nil {
			return e
		}
		old, e := idempotent(tx, u, "invite-accept", key, s.inviteDigest("request", digest(c)))
		if e != nil {
			return e
		}
		if old != 0 {
			out, _, e = familyAccess(tx, u, stringID(old), true)
			return e
		}
		inv, e := s.findInvitation(tx, c)
		if e != nil {
			return e
		}
		var fam h.Family
		if e = tx.Raw("SELECT * FROM household WHERE household_id = ? AND state='active' FOR UPDATE", inv.HouseholdID).Scan(&fam).Error; e != nil {
			return e
		}
		if fam.ID == 0 {
			return inviteUnavailable
		}
		inv, e = s.findInvitation(tx, c)
		if e != nil {
			return e
		}
		var admin h.Member
		if e = tx.Where("member_id = ? AND role='admin' AND state='active'", inv.IssuedBy).Take(&admin).Error; e != nil {
			return inviteUnavailable
		}
		var n int64
		if e = tx.Model(&h.Member{}).Where("user_no = ? AND state='active'", u).Count(&n).Error; e != nil {
			return e
		}
		if n > 0 {
			return h.Fail("ALREADY_IN_HOUSEHOLD", "既に家族へ参加しています")
		}
		var ms []h.Member
		if e = tx.Where("household_id = ? AND state='active'", fam.ID).Find(&ms).Error; e != nil {
			return e
		}
		if len(ms) >= 3 {
			return h.Fail("HOUSEHOLD_FULL", "この家族は満員です")
		}
		used := map[int]bool{}
		for _, m := range ms {
			used[*m.SlotNo] = true
		}
		slot := 1
		for used[slot] {
			slot++
		}
		var m h.Member
		e = tx.Where("household_id = ? AND user_no = ?", fam.ID, u).Take(&m).Error
		if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			return e
		}
		if m.ID == 0 {
			m = h.Member{HouseholdID: fam.ID, UserNo: hid(u), DisplayName: fmt.Sprintf("メンバー%d", slot)}
		}
		m.Role = "member"
		m.State = "active"
		m.SlotNo = &slot
		m.JoinedAt = time.Now()
		m.LeftAt = nil
		if e = tx.Save(&m).Error; e != nil {
			return e
		}
		r := tx.Model(&h.Invitation{}).Where("invitation_id = ? AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP", inv.ID).Updates(map[string]any{"consumed_by": u, "consumed_at": gorm.Expr("CURRENT_TIMESTAMP")})
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return inviteUnavailable
		}
		if e = bumpFamily(tx, stringID(fam.ID)); e != nil {
			return e
		}
		if e = remember(tx, u, "invite-accept", key, s.inviteDigest("request", digest(c)), fam.ID); e != nil {
			return e
		}
		fam.Version++
		fam.Role = m.Role
		fam.MemberID = m.ID
		out = &fam
		return nil
	})
	return out, e
}
