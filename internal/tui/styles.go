package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Monochrome + cyan neon accent palette
	colorAccent  = lipgloss.AdaptiveColor{Light: "#0097A7", Dark: "#00E5FF"}
	colorText    = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#CCCCCC"}
	colorMuted   = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#666666"}
	colorDim     = lipgloss.AdaptiveColor{Light: "#999999", Dark: "#444444"}
	colorSubtle  = lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#222222"}
	colorSurface = lipgloss.AdaptiveColor{Light: "#F0F0F0", Dark: "#111111"}
	colorBody    = lipgloss.AdaptiveColor{Light: "#444444", Dark: "#AAAAAA"}

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

	previewSummaryStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Italic(true)

	previewTagsStyle = lipgloss.NewStyle().
				Foreground(colorDim)

	previewLinkStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Italic(true).
				MarginTop(1)

	tabActiveStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Bold(true)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorText)

	tabSeparatorStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

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

	// Briefing V2 styles — neon palette
	briefingTitleColor = lipgloss.Color("#00FFFF")
	briefingBodyColor  = lipgloss.Color("#E0E0E0")
	briefingMetaColor  = lipgloss.Color("#00E5FF")
	briefingWhyColor   = lipgloss.Color("#B0FFB0")
	briefingV2TitleStyle = lipgloss.NewStyle().
				Foreground(briefingTitleColor).
				Bold(true)

	briefingV2BodyStyle = lipgloss.NewStyle().
				Foreground(briefingBodyColor)

	briefingV2MetaStyle = lipgloss.NewStyle().
				Foreground(briefingMetaColor)

	briefingV2WhyStyle = lipgloss.NewStyle().
				Foreground(briefingWhyColor)

	categoryColors = map[string]lipgloss.Color{
		"AI/ML":               lipgloss.Color("#DA70D6"),
		"Infrastructure":      lipgloss.Color("#00FFAB"),
		"Databases":           lipgloss.Color("#7FFF00"),
		"Distributed Systems": lipgloss.Color("#FFD700"),
		"Security":            lipgloss.Color("#FF6B6B"),
		"Developer Tools":     lipgloss.Color("#87CEEB"),
		"Platform":            lipgloss.Color("#00E5FF"),
	}
)

func categoryStyle(cat string) lipgloss.Style {
	color, ok := categoryColors[cat]
	if !ok {
		color = lipgloss.Color("#80DEEA")
	}
	return lipgloss.NewStyle().Foreground(color)
}
