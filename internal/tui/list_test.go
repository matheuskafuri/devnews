package tui

import (
	"testing"
	"time"
)

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
		{"", 5, ""},
		{"test", 0, ""},
	}
	for _, tt := range tests {
		got := truncateStr(tt.input, tt.n)
		if got != tt.want {
			t.Errorf("truncateStr(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.want)
		}
	}
}

func TestTruncateStrUTF8(t *testing.T) {
	got := truncateStr("日本語テスト", 5)
	want := "日本..."
	if got != want {
		t.Errorf("truncateStr(Japanese, 5) = %q, want %q", got, want)
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		t    time.Time
		want string
	}{
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5m"},
		{now.Add(-3 * time.Hour), "3h"},
		{now.Add(-2 * 24 * time.Hour), "2d"},
	}
	for _, tt := range tests {
		got := relativeTime(tt.t)
		if got != tt.want {
			t.Errorf("relativeTime(%v ago) = %q, want %q", now.Sub(tt.t), got, tt.want)
		}
	}
}

func TestRelativeTimeOld(t *testing.T) {
	old := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	got := relativeTime(old)
	if got != "Jun 15" {
		t.Errorf("relativeTime(old date) = %q, want %q", got, "Jun 15")
	}
}
