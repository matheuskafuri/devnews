package cmd

import (
	"testing"
	"time"
)

func TestParseSince(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
		err   bool
	}{
		{"7d", 7 * 24 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"24h", 24 * time.Hour, false},
		{"30m", 30 * time.Minute, false},
		{"2h30m", 2*time.Hour + 30*time.Minute, false},
		{"invalid", 0, true},
		{"", 0, true},
		{"d", 0, true},
	}

	for _, tt := range tests {
		got, err := parseSince(tt.input)
		if tt.err {
			if err == nil {
				t.Errorf("parseSince(%q): expected error, got %v", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseSince(%q): unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseSince(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
