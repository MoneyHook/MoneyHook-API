package store_postgres

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

func referenceTable(kind string) (string, string) {
	if kind == "payments" {
		return "household_payment", "payment_id"
	}
	return "household_sub_category", "sub_category_id"
}
func (s *HouseholdStore) References(ctx context.Context, u, f, kind string) ([]h.Reference, error) {
	tx := s.db.WithContext(ctx)
	if _, _, e := familyAccess(tx, u, f, false); e != nil {
		return nil, e
	}
	table, col := referenceTable(kind)
	selects := "CAST(" + col + " AS TEXT) AS id,name,active,version"
	if kind == "payments" {
		selects += ",CAST(payment_type_id AS TEXT) AS payment_type_id,payment_date,closing_date"
	} else {
		selects += ",CAST(category_id AS TEXT) AS category_id"
	}
	rows := []h.Reference{}
	e := tx.Table(table).Select(selects).Where("household_id = ?", f).Order(col).Scan(&rows).Error
	return rows, e
}
func (s *HouseholdStore) SaveReference(ctx context.Context, u, f, kind, id string, in h.ReferenceInput) (*h.Reference, error) {
	var resultID string
	e := householdTransaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		if _, _, e := familyAccess(tx, u, f, true); e != nil {
			return e
		}
		table, col := referenceTable(kind)
		values := map[string]any{}
		var previous h.Reference
		if id != "" {
			if e := tx.Table(table).Select(col+" AS id,name,active,version").Where("household_id = ? AND "+col+" = ?", f, id).Take(&previous).Error; e != nil {
				return h.NotFound
			}
			if e := checkVersion(in.ExpectedVersion, previous.Version); e != nil {
				return e
			}
		}
		if in.Name != "" || id == "" {
			max := 32
			if kind != "payments" {
				max = 16
			}
			if !h.Name(in.Name, max) {
				return h.Invalid
			}
			values["name"] = strings.TrimSpace(in.Name)
		}
		if kind == "payments" {
			if in.PaymentTypeID != nil {
				if !h.ID(*in.PaymentTypeID) {
					return h.Invalid
				}
				var n int64
				if e := tx.Table("payment_type").Where("payment_type_id = ?", *in.PaymentTypeID).Count(&n).Error; e != nil {
					return e
				}
				if n != 1 {
					return h.Invalid
				}
				values["payment_type_id"] = *in.PaymentTypeID
			}
			for _, v := range []*int{in.PaymentDate, in.ClosingDate} {
				if v != nil && (*v < 1 || *v > 31) {
					return h.Invalid
				}
			}
			if id == "" || in.PaymentTypeID != nil {
				if in.PaymentTypeID == nil {
					return h.Invalid
				}
				values["payment_date"] = in.PaymentDate
				values["closing_date"] = in.ClosingDate
			}
		} else if in.CategoryID != nil {
			if !h.ID(*in.CategoryID) {
				return h.Invalid
			}
			var n int64
			if e := tx.Table("category").Where("category_id = ?", *in.CategoryID).Count(&n).Error; e != nil {
				return e
			}
			if n != 1 {
				return h.Invalid
			}
			values["category_id"] = *in.CategoryID
		} else if id == "" {
			return h.Invalid
		}
		if id != "" && len(values) > 0 {
			var n int64
			if e := tx.Model(&householdEntryRecord{}).Where("household_id = ? AND "+col+" = ?", f, id).Count(&n).Error; e != nil {
				return e
			}
			if n > 0 {
				return h.Fail("REFERENCE_IN_USE", "使用済みの項目は無効化だけが可能です")
			}
		}
		if in.Active != nil {
			values["active"] = *in.Active
		}
		if id == "" {
			values["household_id"] = f
			values["version"] = 1
			if _, ok := values["active"]; !ok {
				values["active"] = true
			}
			var n uint64
			e := tx.Table(table).Clauses(returnID(col)).Create(values).Error
			if e != nil {
				return e
			}
			switch v := values[col].(type) {
			case int64:
				n = uint64(v)
			case uint64:
				n = v
			}
			if n == 0 {
				return h.Fail("INTERNAL_ERROR", "ID取得に失敗しました")
			}
			resultID = stringID(n)
		} else {
			values["version"] = previous.Version + 1
			if e := tx.Table(table).Where("household_id = ? AND "+col+" = ?", f, id).Updates(values).Error; e != nil {
				return e
			}
			resultID = id
		}
		return nil
	})
	if e != nil {
		return nil, e
	}
	rows, e := s.References(ctx, u, f, kind)
	if e != nil {
		return nil, e
	}
	for _, r := range rows {
		if r.ID == resultID {
			return &r, nil
		}
	}
	return nil, h.NotFound
}

func returnID(column string) clause.Returning {
	return clause.Returning{Columns: []clause.Column{{Name: column}}}
}
