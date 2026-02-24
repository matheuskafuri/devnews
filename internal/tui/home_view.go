package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var asciiLogo = []string{
	`██████╗ ███████╗██╗   ██╗███╗   ██╗███████╗██╗    ██╗███████╗`,
	`██╔══██╗██╔════╝██║   ██║████╗  ██║██╔════╝██║    ██║██╔════╝`,
	`██║  ██║█████╗  ██║   ██║██╔██╗ ██║█████╗  ██║ █╗ ██║███████╗`,
	`██║  ██║██╔══╝  ╚██╗ ██╔╝██║╚██╗██║██╔══╝  ██║███╗██║╚════██║`,
	`██████╔╝███████╗ ╚████╔╝ ██║ ╚████║███████╗╚███╔███╔╝███████║`,
	`╚═════╝ ╚══════╝  ╚═══╝  ╚═╝  ╚═══╝╚══════╝ ╚══╝╚══╝ ╚══════╝`,
}

func renderHomeScreen(width, height int, hasBriefing bool, updateVersion string) string {
	logoStyle := lipgloss.NewStyle().Foreground(colorAccent)
	keyStyle := lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(colorText)

	var lines []string

	// ASCII logo
	for _, l := range asciiLogo {
		lines = append(lines, logoStyle.Render(l))
	}
	lines = append(lines, "")
	lines = append(lines, "")

	// Menu items
	if hasBriefing {
		lines = append(lines, "          "+keyStyle.Render("[b]")+"  "+labelStyle.Render("Today's Briefing"))
	}
	lines = append(lines, "          "+keyStyle.Render("[e]")+"  "+labelStyle.Render("Browse / Explore"))
	lines = append(lines, "")
	lines = append(lines, "          "+keyStyle.Render("[q]")+"  "+labelStyle.Render("Quit"))

	// Update notification
	if updateVersion != "" {
		lines = append(lines, "")
		lines = append(lines, "          "+logoStyle.Render("Update available: v"+updateVersion+" → brew upgrade devnews"))
	}

	content := strings.Join(lines, "\n")
	contentHeight := strings.Count(content, "\n") + 1

	topPad := (height - contentHeight) / 3
	if topPad < 0 {
		topPad = 0
	}

	// Center horizontally
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Top,
		strings.Repeat("\n", topPad)+content)
}
