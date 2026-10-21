package household

import "testing"

func TestValidBulkSharing(t *testing.T) {
	for _, tc := range []struct {
		name  string
		ids   []string
		valid bool
	}{
		{"missing", nil, false}, {"empty", []string{}, false},
		{"valid", []string{"1", "2"}, true}, {"duplicates", []string{"1", "1"}, true},
		{"zero", []string{"0"}, false}, {"negative", []string{"-1"}, false},
		{"nonnumeric", []string{"abc"}, false}, {"overflow", []string{"9223372036854775808"}, false},
		{"limit", repeatedIDs(1000), true}, {"over limit", repeatedIDs(1001), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidBulkSharing(BulkSharingInput{TransactionIDs: tc.ids}); got != tc.valid {
				t.Fatalf("got %v", got)
			}
		})
	}
}
func repeatedIDs(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = "1"
	}
	return ids
}
