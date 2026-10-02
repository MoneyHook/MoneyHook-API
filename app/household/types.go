package household

import (
	"context"
	"time"
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string        { return e.Message }
func Fail(code, message string) error { return &Error{code, message} }

var NotFound = &Error{"NOT_FOUND", "家族または記録が見つかりません"}
var Forbidden = &Error{"FORBIDDEN", "この操作は許可されていません"}
var Conflict = &Error{"VERSION_CONFLICT", "他の操作で更新されました。再読み込みしてください"}
var Invalid = &Error{"VALIDATION_ERROR", "入力内容を確認してください"}

type Family struct {
	ID         uint64     `json:"household_id,string" gorm:"column:household_id;primaryKey;autoIncrement"`
	Name       string     `json:"name"`
	State      string     `json:"state"`
	Version    int64      `json:"version"`
	CreatedAt  time.Time  `json:"created_at"`
	ArchivedAt *time.Time `json:"archived_at"`
	Role       string     `json:"role" gorm:"-"`
	MemberID   uint64     `json:"member_id,string" gorm:"-"`
}

func (Family) TableName() string { return "household" }

type Member struct {
	ID          uint64     `json:"member_id,string" gorm:"column:member_id;primaryKey;autoIncrement"`
	HouseholdID uint64     `json:"household_id,string"`
	UserNo      uint64     `json:"-"`
	Role        string     `json:"role"`
	State       string     `json:"state"`
	SlotNo      *int       `json:"-"`
	DisplayName string     `json:"display_name"`
	JoinedAt    time.Time  `json:"joined_at"`
	LeftAt      *time.Time `json:"left_at"`
}

func (Member) TableName() string { return "household_member" }

type Invitation struct {
	ID          uint64     `json:"invitation_id,string" gorm:"column:invitation_id;primaryKey;autoIncrement"`
	HouseholdID uint64     `json:"-"`
	IssuedBy    uint64     `json:"-"`
	CodeDigest  string     `json:"-"`
	TokenDigest string     `json:"-"`
	ExpiresAt   time.Time  `json:"expires_at"`
	ConsumedBy  *uint64    `json:"-"`
	ConsumedAt  *time.Time `json:"consumed_at"`
	RevokedAt   *time.Time `json:"revoked_at"`
	CreatedAt   time.Time  `json:"created_at"`
	State       string     `json:"state" gorm:"-"`
}

func (Invitation) TableName() string { return "household_invitation" }

type InviteSecret struct {
	Invitation
	Code  string `json:"code"`
	Token string `json:"token"`
}
type Credential struct {
	Code  string `json:"code,omitempty"`
	Token string `json:"token,omitempty"`
}
type Preview struct {
	Name      string    `json:"household_name"`
	Inviter   string    `json:"inviter_name"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Version struct {
	ExpectedVersion int64 `json:"expected_version"`
}
type FamilyInput struct {
	Version
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}
type MemberInput struct {
	Version
	MemberID    string `json:"target_member_id"`
	DisplayName string `json:"display_name"`
}
type Payer struct {
	Kind        string  `json:"kind"`
	MemberID    *string `json:"member_id"`
	DisplayName string  `json:"display_name,omitempty"`
	State       string  `json:"state,omitempty"`
}
type TransactionInput struct {
	Date            string  `json:"transaction_date"`
	Time            *string `json:"transaction_time"`
	Name            string  `json:"transaction_name"`
	Amount          int64   `json:"amount"`
	Sign            int     `json:"sign"`
	CategoryID      string  `json:"category_id"`
	Fixed           bool    `json:"fixed_flg"`
	SubCategoryID   string  `json:"sub_category_id,omitempty"`
	SubCategoryName string  `json:"sub_category_name,omitempty"`
	PaymentID       *string `json:"payment_id"`
}
type EntryInput struct {
	Version
	Transaction   TransactionInput `json:"transaction"`
	Payer         Payer            `json:"payer"`
	PaymentID     *string          `json:"household_payment_id"`
	SubCategoryID *string          `json:"household_sub_category_id"`
	SourceVersion int64            `json:"source_version"`
	Excluded      bool             `json:"excluded_from_totals"`
}
type Permissions struct {
	Edit    bool `json:"can_edit"`
	Delete  bool `json:"can_delete"`
	Unshare bool `json:"can_unshare"`
	Correct bool `json:"can_correct"`
}
type Entry struct {
	ID      string `json:"entry_id"`
	Kind    string `json:"kind"`
	Version int64  `json:"version"`
	TransactionInput
	SignedAmount             int64       `json:"signed_amount"`
	CategoryName             string      `json:"category_name"`
	HouseholdPaymentID       *string     `json:"household_payment_id"`
	HouseholdPaymentName     *string     `json:"household_payment_name"`
	HouseholdSubCategoryID   *string     `json:"household_sub_category_id"`
	HouseholdSubCategoryName *string     `json:"household_sub_category_name"`
	Payer                    Payer       `json:"payer"`
	CreatedBy                string      `json:"created_by"`
	UpdatedBy                string      `json:"updated_by"`
	UpdatedAt                time.Time   `json:"updated_at"`
	SourceTransactionID      *string     `json:"source_transaction_id,omitempty"`
	SourceVersion            *int64      `json:"source_version,omitempty"`
	Permissions              Permissions `json:"permissions"`
	CapturedAt               *time.Time  `json:"captured_at"`
	Corrected                bool        `json:"corrected"`
	Excluded                 bool        `json:"excluded_from_totals"`
}
type EntryPage struct {
	Entries    []Entry `json:"entries"`
	NextCursor *string `json:"next_cursor"`
}
type Filter struct {
	Month  string
	Payer  string
	Kind   string
	Cursor string
}
type Reference struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	CategoryID    *string `json:"category_id"`
	PaymentTypeID *string `json:"payment_type_id"`
	PaymentDate   *int    `json:"payment_date"`
	ClosingDate   *int    `json:"closing_date"`
	Active        bool    `json:"active"`
	Version       int64   `json:"version"`
}
type ReferenceInput struct {
	Version
	Name          string  `json:"name"`
	CategoryID    *string `json:"category_id"`
	PaymentTypeID *string `json:"payment_type_id"`
	PaymentDate   *int    `json:"payment_date"`
	ClosingDate   *int    `json:"closing_date"`
	Active        *bool   `json:"active"`
}
type Summary struct {
	Income  int64          `json:"income"`
	Expense int64          `json:"expense"`
	Balance int64          `json:"balance"`
	Groups  []SummaryGroup `json:"groups"`
}
type SummaryGroup struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Amount int64  `json:"amount"`
}

type ShareStatus struct {
	EntryID string `json:"entry_id,omitempty"`
	Version int64  `json:"version"`
	State   string `json:"state"`
	Kind    string `json:"kind,omitempty"`
}

type Store interface {
	ShareStatus(context.Context, string, string, string) (*ShareStatus, error)
	List(context.Context, string) ([]Family, error)
	Get(context.Context, string, string) (*Family, error)
	Create(context.Context, string, FamilyInput, string) (*Family, error)
	Rename(context.Context, string, string, FamilyInput) (*Family, error)
	Members(context.Context, string, string) ([]Member, error)
	RenameMember(context.Context, string, string, string) error
	Transfer(context.Context, string, string, MemberInput) error
	Leave(context.Context, string, string, string, int64, bool) error
	Invitations(context.Context, string, string) ([]Invitation, error)
	Issue(context.Context, string, string, string, string) (*InviteSecret, error)
	Revoke(context.Context, string, string, string) error
	Preview(context.Context, string, string, Credential) (*Preview, error)
	Accept(context.Context, string, string, Credential, string) (*Family, error)
	Entries(context.Context, string, string, Filter) (*EntryPage, error)
	Entry(context.Context, string, string, string) (*Entry, error)
	CreateEntry(context.Context, string, string, EntryInput, bool, string) (*Entry, error)
	Share(context.Context, string, string, string, EntryInput) (*Entry, error)
	Unshare(context.Context, string, string, string, int64) error
	UpdateProxy(context.Context, string, string, string, EntryInput) (*Entry, error)
	DeleteProxy(context.Context, string, string, string, int64) error
	Correct(context.Context, string, string, string, EntryInput) (*Entry, error)
	References(context.Context, string, string, string) ([]Reference, error)
	SaveReference(context.Context, string, string, string, string, ReferenceInput) (*Reference, error)
}
