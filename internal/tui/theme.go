package tui

import "github.com/charmbracelet/lipgloss"

// Theme holds all color slots for a devnews theme.
type Theme struct {
	Name    string
	Accent  lipgloss.Color
	Text    lipgloss.Color
	Muted   lipgloss.Color
	Dim     lipgloss.Color
	Subtle  lipgloss.Color
	Surface lipgloss.Color
	Body    lipgloss.Color

	// Briefing-specific
	BriefingTitle lipgloss.Color
	BriefingBody  lipgloss.Color
	BriefingMeta  lipgloss.Color
	BriefingWhy   lipgloss.Color

	// Category colors
	CategoryColors  map[string]lipgloss.Color
	CategoryDefault lipgloss.Color
}

var themes = map[string]Theme{
	"neon": {
		Name:    "neon",
		Accent:  lipgloss.Color("#00E5FF"),
		Text:    lipgloss.Color("#CCCCCC"),
		Muted:   lipgloss.Color("#666666"),
		Dim:     lipgloss.Color("#444444"),
		Subtle:  lipgloss.Color("#222222"),
		Surface: lipgloss.Color("#111111"),
		Body:    lipgloss.Color("#AAAAAA"),
		BriefingTitle: lipgloss.Color("#00FFFF"),
		BriefingBody:  lipgloss.Color("#E0E0E0"),
		BriefingMeta:  lipgloss.Color("#00E5FF"),
		BriefingWhy:   lipgloss.Color("#B0FFB0"),
		CategoryColors: map[string]lipgloss.Color{
			"AI/ML":               lipgloss.Color("#DA70D6"),
			"Infrastructure":      lipgloss.Color("#00FFAB"),
			"Databases":           lipgloss.Color("#7FFF00"),
			"Distributed Systems": lipgloss.Color("#FFD700"),
			"Security":            lipgloss.Color("#FF6B6B"),
			"Developer Tools":     lipgloss.Color("#87CEEB"),
			"Platform":            lipgloss.Color("#00E5FF"),
		},
		CategoryDefault: lipgloss.Color("#80DEEA"),
	},
	"dracula": {
		Name:    "dracula",
		Accent:  lipgloss.Color("#BD93F9"),
		Text:    lipgloss.Color("#F8F8F2"),
		Muted:   lipgloss.Color("#6272A4"),
		Dim:     lipgloss.Color("#44475A"),
		Subtle:  lipgloss.Color("#282A36"),
		Surface: lipgloss.Color("#21222C"),
		Body:    lipgloss.Color("#BFBFBF"),
		BriefingTitle: lipgloss.Color("#FF79C6"),
		BriefingBody:  lipgloss.Color("#F8F8F2"),
		BriefingMeta:  lipgloss.Color("#8BE9FD"),
		BriefingWhy:   lipgloss.Color("#50FA7B"),
		CategoryColors: map[string]lipgloss.Color{
			"AI/ML":               lipgloss.Color("#FF79C6"),
			"Infrastructure":      lipgloss.Color("#50FA7B"),
			"Databases":           lipgloss.Color("#F1FA8C"),
			"Distributed Systems": lipgloss.Color("#FFB86C"),
			"Security":            lipgloss.Color("#FF5555"),
			"Developer Tools":     lipgloss.Color("#8BE9FD"),
			"Platform":            lipgloss.Color("#BD93F9"),
		},
		CategoryDefault: lipgloss.Color("#8BE9FD"),
	},
	"nord": {
		Name:    "nord",
		Accent:  lipgloss.Color("#88C0D0"),
		Text:    lipgloss.Color("#ECEFF4"),
		Muted:   lipgloss.Color("#616E88"),
		Dim:     lipgloss.Color("#4C566A"),
		Subtle:  lipgloss.Color("#3B4252"),
		Surface: lipgloss.Color("#2E3440"),
		Body:    lipgloss.Color("#D8DEE9"),
		BriefingTitle: lipgloss.Color("#88C0D0"),
		BriefingBody:  lipgloss.Color("#D8DEE9"),
		BriefingMeta:  lipgloss.Color("#81A1C1"),
		BriefingWhy:   lipgloss.Color("#A3BE8C"),
		CategoryColors: map[string]lipgloss.Color{
			"AI/ML":               lipgloss.Color("#B48EAD"),
			"Infrastructure":      lipgloss.Color("#A3BE8C"),
			"Databases":           lipgloss.Color("#EBCB8B"),
			"Distributed Systems": lipgloss.Color("#D08770"),
			"Security":            lipgloss.Color("#BF616A"),
			"Developer Tools":     lipgloss.Color("#88C0D0"),
			"Platform":            lipgloss.Color("#81A1C1"),
		},
		CategoryDefault: lipgloss.Color("#88C0D0"),
	},
	"solarized-light": {
		Name:    "solarized-light",
		Accent:  lipgloss.Color("#268BD2"),
		Text:    lipgloss.Color("#073642"),
		Muted:   lipgloss.Color("#586E75"),
		Dim:     lipgloss.Color("#93A1A1"),
		Subtle:  lipgloss.Color("#EEE8D5"),
		Surface: lipgloss.Color("#FDF6E3"),
		Body:    lipgloss.Color("#657B83"),
		BriefingTitle: lipgloss.Color("#268BD2"),
		BriefingBody:  lipgloss.Color("#073642"),
		BriefingMeta:  lipgloss.Color("#2AA198"),
		BriefingWhy:   lipgloss.Color("#859900"),
		CategoryColors: map[string]lipgloss.Color{
			"AI/ML":               lipgloss.Color("#D33682"),
			"Infrastructure":      lipgloss.Color("#859900"),
			"Databases":           lipgloss.Color("#B58900"),
			"Distributed Systems": lipgloss.Color("#CB4B16"),
			"Security":            lipgloss.Color("#DC322F"),
			"Developer Tools":     lipgloss.Color("#2AA198"),
			"Platform":            lipgloss.Color("#268BD2"),
		},
		CategoryDefault: lipgloss.Color("#2AA198"),
	},
}

// ThemeNames returns sorted theme names for the picker.
func ThemeNames() []string {
	return []string{"neon", "dracula", "nord", "solarized-light"}
}

// GetTheme returns a theme by name, defaulting to neon.
func GetTheme(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["neon"]
}

// currentTheme is the active theme. Set at startup.
var currentTheme = themes["neon"]

// applyTheme updates all style variables to use the given theme's colors.
func applyTheme(t Theme) {
	currentTheme = t

	colorAccent = t.Accent
	colorText = t.Text
	colorMuted = t.Muted
	colorDim = t.Dim
	colorSubtle = t.Subtle
	colorSurface = t.Surface
	colorBody = t.Body

	// Rebuild all styles with new colors
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent).PaddingLeft(1)
	headerDateStyle = lipgloss.NewStyle().Foreground(colorDim).Align(lipgloss.Right)
	listPaneStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(colorSubtle)
	listPaneActiveStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(colorAccent)
	previewPaneStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(colorSubtle)
	previewPaneActiveStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(colorAccent)
	itemTitleStyle = lipgloss.NewStyle().Foreground(colorText)
	itemSelectedStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	itemUnreadStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	itemReadStyle = lipgloss.NewStyle().Foreground(colorDim)
	itemAIMarkerStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	itemSourceStyle = lipgloss.NewStyle().Foreground(colorMuted)
	itemTimeStyle = lipgloss.NewStyle().Foreground(colorDim)
	previewTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent).MarginBottom(1)
	previewSourceStyle = lipgloss.NewStyle().Foreground(colorMuted)
	previewBodyStyle = lipgloss.NewStyle().Foreground(colorBody)
	previewTagsStyle = lipgloss.NewStyle().Foreground(colorDim)
	previewLinkStyle = lipgloss.NewStyle().Foreground(colorDim).Italic(true).MarginTop(1)
	fullSummaryLabelStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	fullSummaryStyle = lipgloss.NewStyle().Foreground(colorBody).Italic(true)
	statusBarStyle = lipgloss.NewStyle().Background(colorSurface).Foreground(colorMuted).PaddingLeft(1).PaddingRight(1)
	spinnerStyle = lipgloss.NewStyle().Foreground(colorAccent)
	searchPromptStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	helpCardStyle = lipgloss.NewStyle().Foreground(colorText).Border(lipgloss.NormalBorder()).BorderForeground(colorSubtle).Padding(1, 3)
	helpDimStyle = lipgloss.NewStyle().Foreground(colorDim)
	overlayTitleStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	overlayLabelStyle = lipgloss.NewStyle().Foreground(colorText).Bold(true)
	overlayHintStyle = lipgloss.NewStyle().Foreground(colorDim)
	previewRuleStyle = lipgloss.NewStyle().Foreground(colorSubtle)
	previewHintStyle = lipgloss.NewStyle().Foreground(colorDim).MarginTop(1)
	filterOverlayStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorAccent).Padding(1, 2)

	// Briefing styles
	briefingV2TitleStyle = lipgloss.NewStyle().Foreground(t.BriefingTitle).Bold(true)
	briefingV2BodyStyle = lipgloss.NewStyle().Foreground(t.BriefingBody)
	briefingV2MetaStyle = lipgloss.NewStyle().Foreground(t.BriefingMeta)
	briefingV2WhyStyle = lipgloss.NewStyle().Foreground(t.BriefingWhy)

	// Per-render cached styles
	itemTimeFreshStyle = lipgloss.NewStyle().Foreground(colorAccent)

	// Filter overlay styles
	overlayActiveNameStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	overlayInactiveNameStyle = lipgloss.NewStyle().Foreground(colorText)
	overlayCursorStyle = lipgloss.NewStyle().Foreground(colorAccent)
	overlayCheckOnStyle = lipgloss.NewStyle().Foreground(colorAccent)
	overlayCheckOffStyle = lipgloss.NewStyle().Foreground(colorDim)
	overlayHelpStyle = lipgloss.NewStyle().Foreground(colorDim)
}
