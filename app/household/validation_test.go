package household

import "testing"

func TestEntryValidation(t *testing.T) {
	valid := EntryInput{Transaction: TransactionInput{Name: "食費", Date: "2026-09-29", Amount: 1, Sign: -1, CategoryID: "1", SubCategoryName: "外食"}, Payer: Payer{Kind: "common"}}
	for name, change := range map[string]func(*EntryInput){
		"zero":               func(v *EntryInput) { v.Transaction.Amount = 0 },
		"too-large":          func(v *EntryInput) { v.Transaction.Amount = 10000000 },
		"invalid-date":       func(v *EntryInput) { v.Transaction.Date = "2026-02-30" },
		"invalid-time":       func(v *EntryInput) { s := "25:00"; v.Transaction.Time = &s },
		"blank":              func(v *EntryInput) { v.Transaction.Name = "　 " },
		"negative-id":        func(v *EntryInput) { v.Transaction.CategoryID = "-1" },
		"invalid-payer":      func(v *EntryInput) { v.Payer.Kind = "member" },
		"invalid-family-ref": func(v *EntryInput) { s := "bad"; v.PaymentID = &s },
	} {
		t.Run(name, func(t *testing.T) {
			v := valid
			change(&v)
			if ValidEntry(v, false) {
				t.Fatal("accepted invalid input")
			}
		})
	}
	if !ValidEntry(valid, true) || !ValidEntry(valid, false) {
		t.Fatal("valid rejected")
	}
	valid.Transaction.SubCategoryID = "1"
	if ValidEntry(valid, true) {
		t.Fatal("both individual subcategory shapes accepted")
	}
}
