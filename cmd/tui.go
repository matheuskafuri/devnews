package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/matheuskafuri/devnews/internal/ai"
	"github.com/matheuskafuri/devnews/internal/briefing"
	"github.com/matheuskafuri/devnews/internal/cache"
	"github.com/matheuskafuri/devnews/internal/classify"
	"github.com/matheuskafuri/devnews/internal/config"
	"github.com/matheuskafuri/devnews/internal/feed"
	"github.com/matheuskafuri/devnews/internal/tui"
	"github.com/spf13/cobra"
)

func runTUI(cmd *cobra.Command, args []string) error {
	return runApp(false)
}

func runApp(browseMode bool) error {
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

	// Update reading streak
	streak, _ := db.UpdateStreak()

	// Initialize AI summarizer (optional, non-fatal)
	var summarizer ai.Summarizer
	if cfg.AIEnabled() {
		summarizer, _ = ai.New(cfg.AI, cfg.AIKey())
	}

	// Generate V2 briefing (unless browse mode)
	var briefingV2 *briefing.Briefing
	if !browseMode {
		// Resolve focus
		focusCategory := ""
		focus := flagFocus
		if focus == "" {
			focus = cfg.DefaultFocus
		}
		if focus != "" {
			cat, err := classify.ResolveAlias(focus)
			if err != nil {
				return err
			}
			focusCategory = string(cat)
		}

		// Determine since for briefing
		briefingSince := since
		if briefingSince.IsZero() {
			// Default: articles from last 24h
			briefingSince = time.Now().Add(-24 * time.Hour)
		}

		b, err := briefing.Generate(briefing.GenerateOpts{
			DB:            db,
			Since:         briefingSince,
			BriefSize:     cfg.GetBriefSize(),
			FocusCategory: focusCategory,
		})
		if err == nil {
			briefingV2 = b
		}
	}

	db.SetLastOpened()

	return tui.Run(tui.RunOpts{
		Cfg:            cfg,
		DB:             db,
		Since:          since,
		Streak:         streak,
		Summarizer:     summarizer,
		BrowseMode:     browseMode,
		BriefingV2:     briefingV2,
		CurrentVersion: Version(),
	})
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
