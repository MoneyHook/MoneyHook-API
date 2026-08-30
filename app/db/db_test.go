package db

import "testing"

func TestParseSeedDataEnabled(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		exists  bool
		want    bool
		wantErr bool
	}{
		{name: "unset", exists: false},
		{name: "empty", value: "", exists: true},
		{name: "whitespace", value: " \t\n", exists: true},
		{name: "true", value: "true", exists: true, want: true},
		{name: "upper true", value: "TRUE", exists: true, want: true},
		{name: "title true", value: " True ", exists: true, want: true},
		{name: "one", value: "1", exists: true, want: true},
		{name: "false", value: "false", exists: true},
		{name: "zero", value: "0", exists: true},
		{name: "invalid", value: "enabled", exists: true, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseSeedDataEnabled(test.value, test.exists)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseSeedDataEnabled() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Errorf("parseSeedDataEnabled() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestQuotePostgresIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       string
	}{
		{name: "simple", identifier: "moneyhook", want: `"moneyhook"`},
		{name: "double quote", identifier: `money"hook`, want: `"money""hook"`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := quotePostgresIdentifier(test.identifier); got != test.want {
				t.Errorf("quotePostgresIdentifier(%q) = %q, want %q", test.identifier, got, test.want)
			}
		})
	}
}
