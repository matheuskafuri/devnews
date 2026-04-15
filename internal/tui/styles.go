package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color variables — reassigned by applyTheme().
	colorAccent  lipgloss.TerminalColor = lipgloss.Color("#00E5FF")
	colorText    lipgloss.TerminalColor = lipgloss.Color("#CCCCCC")
	colorMuted   lipgloss.TerminalColor = lipgloss.Color("#666666")
	colorDim     lipgloss.TerminalColor = lipgloss.Color("#444444")
	colorSubtle  lipgloss.TerminalColor = lipgloss.Color("#222222")
	colorSurface lipgloss.TerminalColor = lipgloss.Color("#111111")
	colorBody    lipgloss.TerminalColor = lipgloss.Color("#AAAAAA")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			PaddingLeft(1)

	headerDateStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Align(lipgloss.Right)

	listPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorSubtle)

	listPaneActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorAccent)

	previewPaneStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorSubtle)

	previewPaneActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(colorAccent)

	itemTitleStyle = lipgloss.NewStyle().
			Foreground(colorText)

	itemSelectedStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	itemUnreadStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	itemReadStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	itemAIMarkerStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	itemSourceStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	itemTimeStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				MarginBottom(1)

	previewSourceStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	previewBodyStyle = lipgloss.NewStyle().
				Foreground(colorBody)

	previewTagsStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	previewLinkStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Italic(true).
				MarginTop(1)

	fullSummaryLabelStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	fullSummaryStyle = lipgloss.NewStyle().
				Foreground(colorBody).
				Italic(true)

	statusBarStyle = lipgloss.NewStyle().
			Background(colorSurface).
			Foreground(colorMuted).
			PaddingLeft(1).
			PaddingRight(1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	searchPromptStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	helpCardStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorSubtle).
			Padding(1, 3)

	helpDimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Briefing V2 styles — set by applyTheme()
	briefingV2TitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFFF")).
				Bold(true)

	briefingV2BodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E0E0E0"))

	briefingV2MetaStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00E5FF"))

	briefingV2WhyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#B0FFB0"))

	// Shared overlay styles
	overlayTitleStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	overlayLabelStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Bold(true)

	overlayHintStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	previewRuleStyle = lipgloss.NewStyle().
				Foreground(colorSubtle)

	previewHintStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				MarginTop(1)

	filterOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorAccent).
				Padding(1, 2)

	// Per-render cached styles (rebuilt by applyTheme)
	itemTimeFreshStyle = lipgloss.NewStyle().Foreground(colorAccent)
)

func overlayBoxStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 3).
		Width(width)
}

func categoryStyle(cat string) lipgloss.Style {
	color, ok := currentTheme.CategoryColors[cat]
	if !ok {
		color = currentTheme.CategoryDefault
	}
	return lipgloss.NewStyle().Foreground(color)
}
