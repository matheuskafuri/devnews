package scrape

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "hello world", "hello world"},
		{"simple tags", "<p>hello</p>", "hello"},
		{"nested tags", "<div><p>hello <b>world</b></p></div>", "hello world"},
		{"script tags", "<script>var x=1;</script>hello", "hello"},
		{"style tags", "<style>.foo{color:red}</style>hello", "hello"},
		{"br and block tags add space", "<p>one</p><p>two</p>", "one two"},
		{"entities", "hello&amp;world &lt;3", "hello&world <3"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripHTML(tt.input)
			if got != tt.want {
				t.Errorf("StripHTML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><title>Test</title><script>var x=1;</script></head><body><h1>Hello</h1><p>World paragraph.</p></body></html>`))
	}))
	defer srv.Close()

	text, err := Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if text == "" {
		t.Fatal("expected non-empty text")
	}
	if !strings.Contains(text, "Hello") || !strings.Contains(text, "World paragraph") {
		t.Errorf("expected extracted text, got %q", text)
	}
	// Should not contain script content
	if strings.Contains(text, "var x") {
		t.Error("script content should be stripped")
	}
}

func TestFetchTruncates(t *testing.T) {
	// Generate a long page
	long := strings.Repeat("word ", 2000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><body><p>" + long + "</p></body></html>"))
	}))
	defer srv.Close()

	text, err := Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(text) > 4500 {
		t.Errorf("expected truncation around 4000 chars, got %d", len(text))
	}
}

func TestFetchError(t *testing.T) {
	_, err := Fetch("http://localhost:1") // nothing listening
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}
