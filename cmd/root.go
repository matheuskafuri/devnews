package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	flagSince   string
	flagRefresh bool
	flagConfig  string
)

var rootCmd = &cobra.Command{
	Use:   "devnews",
	Short: "TUI engineering blog aggregator",
	Long:  "devnews aggregates engineering blog posts from top tech companies into a clean, hacker-style dashboard.",
	RunE:  runTUI,
}

func init() {
	rootCmd.Flags().StringVar(&flagSince, "since", "", "only show articles from the last duration (e.g., 7d, 24h)")
	rootCmd.Flags().BoolVar(&flagRefresh, "refresh", false, "force refresh feeds before launching")
	rootCmd.Flags().StringVar(&flagConfig, "config", "", "path to config file")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(pruneCmd)
	rootCmd.AddCommand(statsCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("devnews %s (commit: %s, built: %s)\n", version, commit, date)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}
