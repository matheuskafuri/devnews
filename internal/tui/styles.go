package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Adaptive colors for dark/light terminals
	colorPrimary   = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	colorSecondary = lipgloss.AdaptiveColor{Light: "#3D3D3D", Dark: "#ABABAB"}
	colorDim       = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#626262"}
	colorAccent    = lipgloss.AdaptiveColor{Light: "#F25D94", Dark: "#F25D94"}
	colorBorder    = lipgloss.AdaptiveColor{Light: "#DBDBDB", Dark: "#383838"}
	colorActiveBdr = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	colorTabActive = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	colorTabBg     = lipgloss.AdaptiveColor{Light: "#EEEEEE", Dark: "#2A2A3E"}
	colorStatusBg  = lipgloss.AdaptiveColor{Light: "#E8E8E8", Dark: "#16213E"}
	colorStatusFg  = lipgloss.AdaptiveColor{Light: "#3D3D3D", Dark: "#ABABAB"}
	colorGreen     = lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#25D366"}

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			PaddingLeft(1)

	headerDateStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			Align(lipgloss.Right)

	listPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder)

	listPaneActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorActiveBdr)

	previewPaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder)

	previewPaneActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorActiveBdr)

	itemTitleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	itemSelectedStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	itemSourceStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	itemTimeStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	previewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				MarginBottom(1)

	previewSourceStyle = lipgloss.NewStyle().
				Foreground(colorGreen).
				MarginBottom(1)

	previewBodyStyle = lipgloss.NewStyle().
				Foreground(colorSecondary)

	previewLinkStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Italic(true).
				MarginTop(1)

	tabActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorTabActive).
			Padding(0, 1).
			Bold(true)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorSecondary).
				Background(colorTabBg).
				Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Background(colorStatusBg).
			Foreground(colorStatusFg).
			PaddingLeft(1).
			PaddingRight(1)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(colorAccent)

	searchPromptStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

)
