package household

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

func ID(s string) bool { n, e := strconv.ParseUint(s, 10, 63); return e == nil && n > 0 }
func Name(s string, max int) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(s))
	return n > 0 && n <= max
}
func ValidTransaction(t TransactionInput, own bool) bool {
	d, e := time.Parse("2006-01-02", t.Date)
	if e != nil || d.Format("2006-01-02") != t.Date || d.Year() < 1 || !Name(t.Name, 32) || t.Amount < 1 || t.Amount > 9999999 || (t.Sign != 1 && t.Sign != -1) || !ID(t.CategoryID) {
		return false
	}
	if t.Time != nil {
		v, e := time.Parse("15:04", *t.Time)
		if e != nil || v.Format("15:04") != *t.Time {
			return false
		}
	}
	if own && ((!ID(t.SubCategoryID) && !Name(t.SubCategoryName, 16)) || (t.SubCategoryID != "" && t.SubCategoryName != "")) {
		return false
	}
	if t.PaymentID != nil && !ID(*t.PaymentID) {
		return false
	}
	return true
}
func ValidEntry(in EntryInput, own bool) bool {
	if !ValidTransaction(in.Transaction, own) {
		return false
	}
	for _, id := range []*string{in.PaymentID, in.SubCategoryID} {
		if id != nil && !ID(*id) {
			return false
		}
	}
	if own {
		return true
	}
	return (in.Payer.Kind == "common" && in.Payer.MemberID == nil) || (in.Payer.Kind == "member" && in.Payer.MemberID != nil && ID(*in.Payer.MemberID))
}
func NormalizeCode(s string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), "-", ""))
}
func ValidCredential(c Credential) bool {
	return (c.Code != "" && c.Token == "" && len(NormalizeCode(c.Code)) == 10) || (c.Token != "" && c.Code == "" && len(c.Token) == 43)
}
func ValidMonth(s string) bool {
	t, e := time.Parse("2006-01-02", s)
	return e == nil && t.Day() == 1 && t.Format("2006-01-02") == s
}
