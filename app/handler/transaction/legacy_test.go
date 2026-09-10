package transaction

import "testing"

func TestParseFrequentTransactionLimit(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
		valid bool
	}{
		{name: "defaults when omitted", want: 20, valid: true},
		{name: "accepts transaction form limit", value: "20", want: 20, valid: true},
		{name: "accepts CSV import limit", value: "100", want: 100, valid: true},
		{name: "rejects non numeric value", value: "many"},
		{name: "rejects zero", value: "0"},
		{name: "rejects value above maximum", value: "101"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, valid := parseFrequentTransactionLimit(test.value)
			if valid != test.valid {
				t.Fatalf("parseFrequentTransactionLimit(%q) valid = %v, want %v", test.value, valid, test.valid)
			}
			if got != test.want {
				t.Errorf("parseFrequentTransactionLimit(%q) = %d, want %d", test.value, got, test.want)
			}
		})
	}
}
