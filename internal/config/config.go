package config

import (
	"embed"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

//go:embed default_config.yaml
var defaultConfigFS embed.FS

type Source struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	URL     string `yaml:"url"`
	Enabled bool   `yaml:"enabled"`
}

type AIConfig struct {
	Provider string `yaml:"provider"` // "claude" or "openai"
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

type Config struct {
	RefreshInterval string    `yaml:"refresh_interval"`
	Retention       string    `yaml:"retention"`
	BriefSize       int       `yaml:"brief_size,omitempty"`
	DefaultFocus    string    `yaml:"focus,omitempty"`
	Sources         []Source  `yaml:"sources"`
	AI              *AIConfig `yaml:"ai,omitempty"`
}

// AIEnabled returns true if AI is configured with a valid API key.
func (c *Config) AIEnabled() bool {
	if c.AI == nil {
		return false
	}
	key := c.AI.APIKey
	if key == "" {
		key = os.Getenv("DEVNEWS_AI_KEY")
	}
	return key != ""
}

// AIKey returns the resolved API key (config or env var).
func (c *Config) AIKey() string {
	if c.AI != nil && c.AI.APIKey != "" {
		return c.AI.APIKey
	}
	return os.Getenv("DEVNEWS_AI_KEY")
}

func (c *Config) RefreshDuration() time.Duration {
	d, err := time.ParseDuration(c.RefreshInterval)
	if err != nil {
		return 12 * time.Hour
	}
	return d
}

func (c *Config) RetentionDuration() time.Duration {
	if c.Retention == "" {
		return 7 * 24 * time.Hour // default: 90 days
	}
	// Support "Nd" day syntax
	if len(c.Retention) > 1 && c.Retention[len(c.Retention)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(c.Retention, "%dd", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}
	d, err := time.ParseDuration(c.Retention)
	if err != nil {
		return 7 * 24 * time.Hour
	}
	return d
}

func (c *Config) EnabledSources() []Source {
	var out []Source
	for _, s := range c.Sources {
		if s.Enabled {
			out = append(out, s)
		}
	}
	return out
}

func (c *Config) SourceNames() []string {
	var names []string
	for _, s := range c.EnabledSources() {
		names = append(names, s.Name)
	}
	return names
}

// GetBriefSize returns the briefing size, defaulting to 5.
func (c *Config) GetBriefSize() int {
	if c.BriefSize <= 0 {
		return 5
	}
	return c.BriefSize
}

func DefaultConfigPath() string {
	return filepath.Join(xdg.ConfigHome, "devnews", "config.yaml")
}

func CachePath() string {
	return filepath.Join(xdg.CacheHome, "devnews", "devnews.db")
}

func loadDefaults() (*Config, error) {
	data, err := defaultConfigFS.ReadFile("default_config.yaml")
	if err != nil {
		return nil, fmt.Errorf("reading embedded config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing embedded config: %w", err)
	}
	return &cfg, nil
}

func Load(path string) (*Config, error) {
	defaults, err := loadDefaults()
	if err != nil {
		return nil, err
	}

	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Write defaults to config path on first run
			if err := writeDefaults(path); err != nil {
				// Non-fatal: just use embedded defaults
				return defaults, nil
			}
			return defaults, nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func writeDefaults(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, _ := defaultConfigFS.ReadFile("default_config.yaml")
	return os.WriteFile(path, data, 0o644)
}

func validate(cfg *Config) error {
	validTypes := map[string]bool{"rss": true, "atom": true}
	for i, s := range cfg.Sources {
		if s.Name == "" {
			return fmt.Errorf("source %d: name is required", i)
		}
		if s.URL == "" {
			return fmt.Errorf("source %q: url is required", s.Name)
		}
		u, err := url.Parse(s.URL)
		if err != nil {
			return fmt.Errorf("source %q: invalid url: %w", s.Name, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("source %q: url scheme must be http or https, got %q", s.Name, u.Scheme)
		}
		if !validTypes[s.Type] {
			return fmt.Errorf("source %q: unknown type %q (valid: rss, atom)", s.Name, s.Type)
		}
	}
	return nil
}
