package update

import "testing"

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"0.4.1", "0.4.0", true},
		{"0.4.1", "0.3.0", true},
		{"1.0.0", "0.9.9", true},
		{"0.4.1", "0.4.1", false},
		{"0.3.0", "0.4.1", false},
		{"0.4.0", "0.4.1", false},
		{"invalid", "0.4.1", true}, // fallback to string inequality
	}
	for _, tt := range tests {
		got := isNewer(tt.latest, tt.current)
		if got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}
