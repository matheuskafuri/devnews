package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func renderStatusBar(articleCount int, filterLabel string, width int, searching bool, refreshing bool) string {
	left := fmt.Sprintf(" %d articles", articleCount)
	if filterLabel != "All" {
		left += " · " + filterLabel
	}

	right := " ↑↓ navigate  o open  / search  f filter  r refresh  ? help  q quit "

	if searching {
		right = " esc cancel  enter search "
	}
	if refreshing {
		left += " (refreshing...)"
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	bar := left + fmt.Sprintf("%*s", gap, "") + right

	return statusBarStyle.Width(width).Render(bar)
}
