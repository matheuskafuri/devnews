package feed

import "testing"

func TestArticleID(t *testing.T) {
	id1 := articleID("https://example.com/post-1")
	id2 := articleID("https://example.com/post-2")
	id1again := articleID("https://example.com/post-1")

	if id1 == id2 {
		t.Error("different URLs should produce different IDs")
	}
	if id1 != id1again {
		t.Error("same URL should produce same ID")
	}
	if len(id1) != 32 {
		t.Errorf("expected 32-char hex string, got %d chars: %s", len(id1), id1)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		n     int
		want  string
	}{
		{"short", 10, "short"},
		{"exactly ten", 11, "exactly ten"},
		{"this is a long string", 10, "this is..."},
		{"abc", 3, "abc"},
		{"abcd", 3, "abc"},
		{"", 5, ""},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.n)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.n, got, tt.want)
		}
	}
}

func TestTruncateUTF8(t *testing.T) {
	// Japanese characters are multi-byte but should truncate by rune
	input := "こんにちは世界です"
	got := truncate(input, 5)
	want := "こん..."
	if got != want {
		t.Errorf("truncate(%q, 5) = %q, want %q", input, got, want)
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<p>Hello</p>", "Hello"},
		{"<b>Bold</b> and <i>italic</i>", "Bold and italic"},
		{"No tags here", "No tags here"},
		{"<div>  Multiple   spaces  </div>", "Multiple spaces"},
		{"", ""},
		{"<a href=\"url\">Link</a> text", "Link text"},
	}
	for _, tt := range tests {
		got := stripHTML(tt.input)
		if got != tt.want {
			t.Errorf("stripHTML(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
