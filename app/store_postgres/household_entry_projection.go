package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

// Snapshot corrections replace the captured transaction; other entries use their original payload.
const householdEntryPayloadExpr = "CASE WHEN e.kind='snapshot' THEN COALESCE(d.corrected_payload,d.payload) ELSE d.payload END"
const householdEntryDateExpr = "CASE WHEN e.kind='shared' THEN TO_CHAR(t.transaction_date,'YYYY-MM-DD') ELSE (" + householdEntryPayloadExpr + ")->>'transaction_date' END"

type householdEntryReadRecord struct {
	Record               householdEntryRecord `gorm:"embedded"`
	OriginalID           *uint64
	OriginalName         string
	OriginalAmount       int64
	OriginalDate         string
	OriginalTime         *string
	OriginalCategoryID   string
	OriginalCategoryName *string
	OriginalFixed        bool
	OriginalOwner        string
	OriginalVersion      int64
	Payload              json.RawMessage
	CorrectedPayload     json.RawMessage
	CapturedAt           *time.Time
	PayerName            *string
	PayerState           string
	PaymentName          string
	SubCategoryName      string
}

func householdEntryReadQuery(tx *gorm.DB) *gorm.DB {
	return householdEntryJoinedQuery(tx).Select(`e.*,
 t.transaction_id AS original_id, t.transaction_name AS original_name,
 t.transaction_amount AS original_amount, TO_CHAR(t.transaction_date,'YYYY-MM-DD') AS original_date,
 LEFT(CAST(t.transaction_time AS TEXT),5) AS original_time,
 CAST(t.category_id AS TEXT) AS original_category_id, c.category_name AS original_category_name,
 t.fixed_flg AS original_fixed, CAST(t.user_no AS TEXT) AS original_owner, t.version AS original_version,
 d.payload, d.corrected_payload, d.captured_at,
 m.display_name AS payer_name, m.state AS payer_state,
 hp.name AS payment_name, hsc.name AS sub_category_name`)
}

func householdEntryJoinedQuery(tx *gorm.DB) *gorm.DB {
	return tx.Table("household_entry e").
		Joins(`LEFT JOIN "transaction" t ON e.kind='shared' AND t.transaction_id=e.source_transaction_id AND t.deleted_at IS NULL`).
		Joins("LEFT JOIN category c ON c.category_id=t.category_id").
		Joins("LEFT JOIN household_entry_data d ON d.entry_id=e.entry_id").
		Joins("LEFT JOIN household_member m ON m.member_id=e.payer_member_id").
		Joins("LEFT JOIN household_payment hp ON hp.payment_id=e.payment_id").
		Joins("LEFT JOIN household_sub_category hsc ON hsc.sub_category_id=e.sub_category_id")
}

func projectHouseholdEntry(r *householdEntryReadRecord, viewer *h.Member) (*h.Entry, error) {
	var out h.Entry
	if r.Record.Kind == "shared" {
		if r.OriginalID == nil || r.OriginalCategoryName == nil {
			return nil, h.NotFound
		}
		out.TransactionInput = h.TransactionInput{Name: r.OriginalName, Date: r.OriginalDate, Time: r.OriginalTime, CategoryID: r.OriginalCategoryID, Fixed: r.OriginalFixed, Amount: r.OriginalAmount, Sign: 1}
		if r.OriginalAmount < 0 {
			out.Amount = -r.OriginalAmount
			out.Sign = -1
		}
		out.SignedAmount = r.OriginalAmount
		out.CategoryName = *r.OriginalCategoryName
		if viewer != nil && stringID(viewer.UserNo) == r.OriginalOwner {
			id := stringID(*r.Record.SourceTransactionID)
			out.SourceTransactionID = &id
			out.SourceVersion = &r.OriginalVersion
			out.Permissions = h.Permissions{Edit: viewer.State == "active", Delete: viewer.State == "active", Unshare: viewer.State == "active"}
		}
	} else {
		if len(r.Payload) == 0 {
			return nil, gorm.ErrRecordNotFound
		}
		if err := json.Unmarshal(r.Payload, &out); err != nil {
			return nil, err
		}
		out.CapturedAt = r.CapturedAt
		if r.Record.Kind == "snapshot" && len(r.CorrectedPayload) > 0 {
			if err := json.Unmarshal(r.CorrectedPayload, &out); err != nil {
				return nil, err
			}
			out.CapturedAt = r.CapturedAt
			out.Corrected = true
		}
		out.Permissions = h.Permissions{Edit: r.Record.Kind == "proxy" && viewer != nil && viewer.State == "active", Delete: r.Record.Kind == "proxy" && viewer != nil && viewer.State == "active", Correct: r.Record.Kind == "snapshot" && viewer != nil && viewer.State == "active"}
	}
	out.ID = stringID(r.Record.ID)
	out.Kind = r.Record.Kind
	out.Version = r.Record.Version
	out.CreatedBy = stringID(r.Record.CreatedBy)
	out.UpdatedBy = stringID(r.Record.UpdatedBy)
	out.UpdatedAt = r.Record.UpdatedAt
	if r.Record.Kind != "snapshot" {
		out.Payer = h.Payer{Kind: "common", DisplayName: "家族共通"}
		if r.Record.PayerMemberID != nil {
			if r.PayerName == nil {
				return nil, gorm.ErrRecordNotFound
			}
			id := stringID(*r.Record.PayerMemberID)
			out.Payer = h.Payer{Kind: "member", MemberID: &id, DisplayName: *r.PayerName, State: r.PayerState}
		}
		out.HouseholdPaymentID = nil
		out.HouseholdPaymentName = nil
		if r.Record.PaymentID != nil {
			id := stringID(*r.Record.PaymentID)
			out.HouseholdPaymentID = &id
			out.HouseholdPaymentName = &r.PaymentName
		}
		out.HouseholdSubCategoryID = nil
		out.HouseholdSubCategoryName = nil
		if r.Record.SubCategoryID != nil {
			id := stringID(*r.Record.SubCategoryID)
			out.HouseholdSubCategoryID = &id
			out.HouseholdSubCategoryName = &r.SubCategoryName
		}
	}
	// Personal references must never be exposed to other household members.
	out.PaymentID = nil
	out.SubCategoryID = ""
	out.SubCategoryName = ""
	return &out, nil
}
