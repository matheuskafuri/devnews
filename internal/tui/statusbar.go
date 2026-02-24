package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func renderStatusBar(articleCount int, filterLabel string, streak int, width int, searching bool, refreshing bool) string {
	streakAccentStyle := lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true)

	left := fmt.Sprintf(" %d articles", articleCount)
	if filterLabel != "All" {
		left += " · " + filterLabel
	}
	if streak >= 1 {
		left += fmt.Sprintf(" · %s %dd", streakAccentStyle.Render("streak"), streak)
	}

	right := " h home  / search  f filter  q quit "

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

func renderBottomBar(streak int, hints string, width int) string {
	streakAccentStyle := lipgloss.NewStyle().
		Foreground(colorAccent).
		Bold(true)

	left := ""
	if streak >= 1 {
		left = fmt.Sprintf(" %s %dd", streakAccentStyle.Render("streak"), streak)
	}

	right := " " + hints + " "

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	bar := left + fmt.Sprintf("%*s", gap, "") + right

	return statusBarStyle.Width(width).Render(bar)
}
