package ai

import (
	"testing"
)

func TestParseSummaryResponse(t *testing.T) {
	input := `SUMMARY: Cloudflare rewrote their DNS proxy in Rust for better performance and memory safety.
TAGS: rust, infrastructure, performance`

	result := parseSummaryResponse(input)

	if result.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if len(result.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d: %v", len(result.Tags), result.Tags)
	}
	if result.Tags[0] != "rust" {
		t.Errorf("expected first tag 'rust', got %q", result.Tags[0])
	}
}

func TestParseSummaryResponseExtraTags(t *testing.T) {
	input := `SUMMARY: Short summary.
TAGS: a, b, c, d, e`

	result := parseSummaryResponse(input)
	if len(result.Tags) != 3 {
		t.Errorf("expected tags capped at 3, got %d", len(result.Tags))
	}
}

func TestParseSummaryResponseEmpty(t *testing.T) {
	result := parseSummaryResponse("")
	if result.Summary != "" {
		t.Errorf("expected empty summary, got %q", result.Summary)
	}
	if len(result.Tags) != 0 {
		t.Errorf("expected no tags, got %v", result.Tags)
	}
}

func TestParseSummaryResponseMalformed(t *testing.T) {
	// No structured format, just plain text
	result := parseSummaryResponse("This is just a random response without format")
	if result.Summary != "" {
		t.Errorf("expected empty summary for malformed input, got %q", result.Summary)
	}
}

func TestParseThemesPlainLines(t *testing.T) {
	input := "Reliability under scale\nAI inference performance\nJVM memory tuning"
	themes := parseThemes(input)
	if len(themes) != 3 {
		t.Fatalf("expected 3 themes, got %d: %v", len(themes), themes)
	}
	if themes[0] != "Reliability under scale" {
		t.Errorf("unexpected first theme: %q", themes[0])
	}
}

func TestParseThemesBulletedList(t *testing.T) {
	input := "• Reliability under scale\n• AI inference performance\n• JVM memory tuning"
	themes := parseThemes(input)
	if len(themes) != 3 {
		t.Fatalf("expected 3 themes, got %d: %v", len(themes), themes)
	}
	if themes[0] != "Reliability under scale" {
		t.Errorf("unexpected first theme: %q", themes[0])
	}
}

func TestParseThemesNumberedList(t *testing.T) {
	input := "1. Reliability under scale\n2. AI inference performance\n3. JVM memory tuning"
	themes := parseThemes(input)
	if len(themes) != 3 {
		t.Fatalf("expected 3 themes, got %d: %v", len(themes), themes)
	}
	if themes[0] != "Reliability under scale" {
		t.Errorf("unexpected first theme: %q", themes[0])
	}
}

func TestParseThemesCapsAtFour(t *testing.T) {
	input := "Theme 1\nTheme 2\nTheme 3\nTheme 4\nTheme 5"
	themes := parseThemes(input)
	if len(themes) != 4 {
		t.Errorf("expected max 4 themes, got %d", len(themes))
	}
}

func TestParseThemesEmpty(t *testing.T) {
	themes := parseThemes("")
	if len(themes) != 0 {
		t.Errorf("expected 0 themes for empty input, got %d", len(themes))
	}
}
