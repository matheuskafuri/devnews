package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/matheuskafuri/devnews/internal/cache"
	"github.com/matheuskafuri/devnews/internal/config"
	"github.com/matheuskafuri/devnews/internal/feed"
	"github.com/matheuskafuri/devnews/internal/tui"
	"github.com/spf13/cobra"
)

func runTUI(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(flagConfig)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	db, err := cache.Open(config.CachePath())
	if err != nil {
		return fmt.Errorf("opening cache: %w", err)
	}
	defer db.Close()

	// Refresh if needed
	if flagRefresh || db.NeedsRefresh(cfg.RefreshDuration()) {
		fmt.Println("Fetching feeds...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		result := feed.FetchAll(ctx, cfg.EnabledSources())
		cancel()

		for _, e := range result.Errors {
			fmt.Printf("  [warn] %v\n", e)
		}

		if err := db.UpsertArticles(result.Articles); err != nil {
			return fmt.Errorf("caching articles: %w", err)
		}
		db.SetLastRefresh()

		// Auto-prune old articles after refresh
		db.Prune(cfg.RetentionDuration())
	}

	// Parse --since
	var since time.Time
	if flagSince != "" {
		d, err := parseSince(flagSince)
		if err != nil {
			return fmt.Errorf("invalid --since value: %w", err)
		}
		since = time.Now().Add(-d)
	}

	return tui.Run(cfg, db, since)
}

func parseSince(s string) (time.Duration, error) {
	if len(s) > 1 && s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err == nil {
			return time.Duration(days) * 24 * time.Hour, nil
		}
	}
	return time.ParseDuration(s)
}
