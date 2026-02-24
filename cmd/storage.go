package cmd

import (
	"fmt"

	"github.com/matheuskafuri/devnews/internal/cache"
	"github.com/matheuskafuri/devnews/internal/config"
	"github.com/spf13/cobra"
)

var flagPruneOlderThan string

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove old articles from the local cache",
	Long: `Delete cached articles older than the retention period and reclaim disk space.

Uses the retention value from config (default: 90d) unless overridden with --older-than.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(flagConfig)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		db, err := cache.Open(config.CachePath())
		if err != nil {
			return fmt.Errorf("opening cache: %w", err)
		}
		defer db.Close()

		retention := cfg.RetentionDuration()
		if flagPruneOlderThan != "" {
			d, err := parseSince(flagPruneOlderThan)
			if err != nil {
				return fmt.Errorf("invalid --older-than value: %w", err)
			}
			retention = d
		}

		deleted, err := db.Prune(retention)
		if err != nil {
			return fmt.Errorf("pruning: %w", err)
		}

		if deleted == 0 {
			fmt.Println("Nothing to prune.")
		} else {
			fmt.Printf("Pruned %d article(s) older than %s.\n", deleted, formatDuration(retention))
		}
		return nil
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath := config.CachePath()
		db, err := cache.Open(dbPath)
		if err != nil {
			return fmt.Errorf("opening cache: %w", err)
		}
		defer db.Close()

		count, size, err := db.Stats(dbPath)
		if err != nil {
			return fmt.Errorf("reading stats: %w", err)
		}

		fmt.Printf("Cache: %s\n", dbPath)
		fmt.Printf("Articles: %d\n", count)
		fmt.Printf("Size: %s\n", formatBytes(size))
		return nil
	},
}

func init() {
	pruneCmd.Flags().StringVar(&flagPruneOlderThan, "older-than", "", "override retention period (e.g., 30d, 720h)")
}

func formatDuration(d interface{ Hours() float64 }) string {
	h := d.(interface{ Hours() float64 }).Hours()
	days := int(h / 24)
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}
	return fmt.Sprintf("%dh", int(h))
}

func formatBytes(b int64) string {
	switch {
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
