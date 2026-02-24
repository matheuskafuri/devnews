package briefing

import (
	"strings"
	"testing"
	"time"

	"github.com/matheuskafuri/devnews/internal/cache"
)

func TestGreeting(t *testing.T) {
	tests := []struct {
		hour     int
		expected string
	}{
		{8, "Good morning"},
		{14, "Good afternoon"},
		{20, "Good evening"},
		{0, "Good morning"},
		{11, "Good morning"},
		{12, "Good afternoon"},
		{17, "Good evening"},
	}

	for _, tt := range tests {
		now := time.Date(2026, 1, 1, tt.hour, 0, 0, 0, time.Local)
		got := greeting(now)
		if got != tt.expected {
			t.Errorf("hour %d: expected %q, got %q", tt.hour, tt.expected, got)
		}
	}
}

func TestActiveSources(t *testing.T) {
	articles := []cache.Article{
		{Source: "Cloudflare"},
		{Source: "Cloudflare"},
		{Source: "Cloudflare"},
		{Source: "GitHub"},
		{Source: "GitHub"},
		{Source: "Stripe"},
	}

	got := activeSources(articles)
	if !strings.HasPrefix(got, "Cloudflare (3)") {
		t.Errorf("expected Cloudflare first, got %q", got)
	}
	if !strings.Contains(got, "GitHub (2)") {
		t.Errorf("expected GitHub (2) in result, got %q", got)
	}
}

func TestActiveSourcesLimitedToThree(t *testing.T) {
	articles := []cache.Article{
		{Source: "A"}, {Source: "B"}, {Source: "C"}, {Source: "D"},
	}

	got := activeSources(articles)
	parts := strings.Split(got, ", ")
	if len(parts) > 3 {
		t.Errorf("expected at most 3 sources, got %d: %q", len(parts), got)
	}
}

func TestTokenize(t *testing.T) {
	tokens := tokenize("Building DNS in Rust for improved performance!")
	found := map[string]bool{}
	for _, tok := range tokens {
		found[tok] = true
	}
	if !found["building"] {
		t.Error("expected 'building' in tokens")
	}
	if !found["rust"] {
		t.Error("expected 'rust' in tokens")
	}
	if found["dns"] {
		t.Error("'dns' should be filtered (< 4 chars)")
	}
	if found["for"] {
		t.Error("'for' should be filtered (stop word)")
	}
}

func TestGenerateLegacy(t *testing.T) {
	now := time.Now()
	newArticles := []cache.Article{
		{Source: "Cloudflare", Title: "Building DNS in Rust", Published: now},
		{Source: "Cloudflare", Title: "Rust performance tuning", Published: now},
		{Source: "GitHub", Title: "Scaling infrastructure with Rust", Published: now},
	}
	allArticles := append(newArticles, cache.Article{
		Source: "Stripe", Title: "Payment processing at scale", Published: now.Add(-24 * time.Hour),
	})

	b := GenerateLegacy(newArticles, allArticles)
	if b.NewCount != 3 {
		t.Errorf("expected NewCount 3, got %d", b.NewCount)
	}
	if b.Greeting == "" {
		t.Error("expected non-empty greeting")
	}
	if b.ActiveSources == "" {
		t.Error("expected non-empty active sources")
	}
	if !strings.Contains(b.Trending, "rust") {
		t.Errorf("expected 'rust' in trending, got %q", b.Trending)
	}
}

func TestGenerateLegacyNoArticles(t *testing.T) {
	b := GenerateLegacy(nil, nil)
	if b.NewCount != 0 {
		t.Errorf("expected 0 new count, got %d", b.NewCount)
	}
	if b.ActiveSources != "" {
		t.Errorf("expected empty active sources, got %q", b.ActiveSources)
	}
}

func TestEstimateReadTime(t *testing.T) {
	// 100 words * 3 / 200 = 1.5 → 1
	short := estimateReadTime(nWords(100))
	if short != 1 {
		t.Errorf("expected 1 min for 100 words, got %d", short)
	}

	// 500 words * 3 / 200 = 7.5 → 7
	long := estimateReadTime(nWords(500))
	if long < 5 {
		t.Errorf("expected >= 5 min for 500 words, got %d", long)
	}

	// Empty
	empty := estimateReadTime("")
	if empty != 1 {
		t.Errorf("expected min 1 for empty, got %d", empty)
	}
}

func TestDescriptionExcerpt(t *testing.T) {
	got := DescriptionExcerpt("This is a complete sentence about infrastructure. And more text follows.")
	if !strings.HasSuffix(got, ".") {
		t.Errorf("expected excerpt to end with period, got %q", got)
	}

	got = DescriptionExcerpt("")
	if got != "" {
		t.Errorf("expected empty for empty input, got %q", got)
	}
}

func nWords(n int) string {
	words := make([]string, n)
	for i := range words {
		words[i] = "word"
	}
	return strings.Join(words, " ")
}
