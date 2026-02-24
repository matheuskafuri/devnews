package cmd

import "github.com/spf13/cobra"

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Launch the full article browser",
	Long:  "Open devnews in browse mode — the classic two-pane article browser.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApp(true)
	},
}
