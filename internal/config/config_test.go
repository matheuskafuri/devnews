package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := loadDefaults()
	if err != nil {
		t.Fatalf("loadDefaults: %v", err)
	}
	if len(cfg.Sources) == 0 {
		t.Error("expected at least one default source")
	}
	if cfg.RefreshInterval == "" {
		t.Error("expected refresh_interval to be set")
	}
}

func TestRefreshDuration(t *testing.T) {
	cfg := &Config{RefreshInterval: "30m"}
	d := cfg.RefreshDuration()
	if d.Minutes() != 30 {
		t.Errorf("expected 30m, got %v", d)
	}

	cfg.RefreshInterval = "invalid"
	d = cfg.RefreshDuration()
	if d.Hours() != 1 {
		t.Errorf("expected 1h default for invalid interval, got %v", d)
	}
}

func TestRetentionDuration(t *testing.T) {
	tests := []struct {
		input    string
		wantDays int
	}{
		{"90d", 90},
		{"30d", 30},
		{"720h", 30},
		{"", 90},       // default
		{"invalid", 90}, // fallback to default
	}
	for _, tt := range tests {
		cfg := &Config{Retention: tt.input}
		got := cfg.RetentionDuration()
		wantHours := float64(tt.wantDays * 24)
		if got.Hours() != wantHours {
			t.Errorf("RetentionDuration(%q) = %v, want %dd", tt.input, got, tt.wantDays)
		}
	}
}

func TestEnabledSources(t *testing.T) {
	cfg := &Config{
		Sources: []Source{
			{Name: "A", Enabled: true},
			{Name: "B", Enabled: false},
			{Name: "C", Enabled: true},
		},
	}
	enabled := cfg.EnabledSources()
	if len(enabled) != 2 {
		t.Fatalf("expected 2 enabled sources, got %d", len(enabled))
	}
	if enabled[0].Name != "A" || enabled[1].Name != "C" {
		t.Errorf("unexpected enabled sources: %v", enabled)
	}
}

func TestSourceNames(t *testing.T) {
	cfg := &Config{
		Sources: []Source{
			{Name: "Alpha", Enabled: true},
			{Name: "Beta", Enabled: false},
			{Name: "Gamma", Enabled: true},
		},
	}
	names := cfg.SourceNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	if names[0] != "Alpha" || names[1] != "Gamma" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `refresh_interval: 2h
sources:
  - name: Test
    type: rss
    url: https://example.com/feed
    enabled: true
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RefreshInterval != "2h" {
		t.Errorf("expected 2h, got %s", cfg.RefreshInterval)
	}
	if len(cfg.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.Sources))
	}
	if cfg.Sources[0].Name != "Test" {
		t.Errorf("expected source name Test, got %s", cfg.Sources[0].Name)
	}
}

func TestLoadNonexistentFallsBackToDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "sub", "config.yaml")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(cfg.Sources) == 0 {
		t.Error("expected default sources when config doesn't exist")
	}
}

func TestValidateMissingName(t *testing.T) {
	cfg := &Config{Sources: []Source{{Type: "rss", URL: "https://example.com"}}}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestValidateMissingURL(t *testing.T) {
	cfg := &Config{Sources: []Source{{Name: "Test", Type: "rss"}}}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestValidateInvalidType(t *testing.T) {
	cfg := &Config{Sources: []Source{{Name: "Test", Type: "json", URL: "https://example.com"}}}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestValidateInvalidURLScheme(t *testing.T) {
	cfg := &Config{Sources: []Source{{Name: "Test", Type: "rss", URL: "file:///etc/passwd"}}}
	err := validate(cfg)
	if err == nil {
		t.Error("expected error for file:// URL scheme")
	}
}

func TestValidateAcceptsHTTPS(t *testing.T) {
	cfg := &Config{Sources: []Source{{Name: "Test", Type: "rss", URL: "https://example.com/feed"}}}
	err := validate(cfg)
	if err != nil {
		t.Errorf("unexpected error for https URL: %v", err)
	}
}

func TestValidateAcceptsHTTP(t *testing.T) {
	cfg := &Config{Sources: []Source{{Name: "Test", Type: "rss", URL: "http://example.com/feed"}}}
	err := validate(cfg)
	if err != nil {
		t.Errorf("unexpected error for http URL: %v", err)
	}
}
