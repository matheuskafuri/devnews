package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matheuskafuri/devnews/internal/config"
)

func (a *App) handleThemePickerKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	names := ThemeNames()

	switch msg.String() {
	case "esc", "T":
		applyTheme(GetTheme(a.originalTheme))
		a.mode = modeNormal
		return a, nil
	case "j", "down":
		if a.themeCursor < len(names)-1 {
			a.themeCursor++
			applyTheme(GetTheme(names[a.themeCursor]))
		}
		return a, nil
	case "k", "up":
		if a.themeCursor > 0 {
			a.themeCursor--
			applyTheme(GetTheme(names[a.themeCursor]))
		}
		return a, nil
	case "enter":
		selected := names[a.themeCursor]
		applyTheme(GetTheme(selected))
		a.mode = modeNormal
		// Persist to config
		return a, func() tea.Msg {
			config.SaveTheme(selected)
			return nil
		}
	}
	return a, nil
}

func (a *App) renderThemePickerOverlay() string {
	names := ThemeNames()

	var b strings.Builder
	b.WriteString(overlayTitleStyle.Render("Select Theme"))
	b.WriteString("\n\n")

	for i, name := range names {
		theme := GetTheme(name)
		// Show a color swatch: accent color block + theme name
		swatch := lipgloss.NewStyle().Foreground(theme.Accent).Render("██")

		if i == a.themeCursor {
			label := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render(name)
			b.WriteString("  ▸ " + swatch + " " + label)
		} else {
			label := lipgloss.NewStyle().Foreground(colorText).Render(name)
			b.WriteString("    " + swatch + " " + label)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(overlayHintStyle.Render("j/k navigate  enter select  esc cancel"))

	return overlayBoxStyle(40).Render(b.String())
}
