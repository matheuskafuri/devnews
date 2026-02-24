package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type filterBar struct {
	sources      []string
	active       map[string]bool
	filterMode   bool
	filterCursor int
}

func newFilterBar(sources []string) filterBar {
	return filterBar{
		sources: sources,
		active:  make(map[string]bool),
	}
}

func (f *filterBar) toggle(source string) {
	if f.active[source] {
		delete(f.active, source)
	} else {
		f.active[source] = true
	}
}

func (f *filterBar) toggleCurrent() {
	if f.filterCursor < len(f.sources) {
		f.toggle(f.sources[f.filterCursor])
	}
}

func (f *filterBar) activeSources() []string {
	if len(f.active) == 0 {
		return nil // nil = all sources
	}
	var out []string
	for _, s := range f.sources {
		if f.active[s] {
			out = append(out, s)
		}
	}
	return out
}

func (f *filterBar) activeLabel() string {
	active := f.activeSources()
	if active == nil {
		return "All"
	}
	return strings.Join(active, ", ")
}

func (f *filterBar) render(width int) string {
	sep := tabSeparatorStyle.Render(" · ")
	arrow := tabSeparatorStyle.Render(" ‹ ")
	arrowR := tabSeparatorStyle.Render(" › ")

	// Build all styled parts: [All, source0, source1, ...]
	type part struct {
		text  string
		width int
	}
	var parts []part

	// "All" tab
	allLabel := "All"
	if len(f.active) == 0 {
		allLabel = tabActiveStyle.Render(allLabel)
	} else {
		allLabel = tabInactiveStyle.Render(allLabel)
	}
	parts = append(parts, part{allLabel, lipgloss.Width(allLabel)})

	for i, s := range f.sources {
		style := tabInactiveStyle
		if f.active[s] {
			style = tabActiveStyle
		}
		label := s
		if f.filterMode && i == f.filterCursor {
			label = "[" + s + "]"
		}
		rendered := style.Render(label)
		parts = append(parts, part{rendered, lipgloss.Width(rendered)})
	}

	sepWidth := lipgloss.Width(sep)
	arrowWidth := lipgloss.Width(arrow)

	// Determine visible window: start from the cursor (index+1 because of "All" tab)
	// and expand to fill width, ensuring the cursor is always visible.
	cursorIdx := f.filterCursor + 1 // offset by "All" tab
	if !f.filterMode {
		cursorIdx = 0
	}

	// Always try to show from the start; scroll right only if needed
	startIdx := 0
	usedWidth := 0
	// Check if cursor fits when rendering from start
	for i := 0; i <= cursorIdx && i < len(parts); i++ {
		if i > 0 {
			usedWidth += sepWidth
		}
		usedWidth += parts[i].width
	}
	// If cursor doesn't fit, find a start that shows it
	if usedWidth > width-arrowWidth {
		startIdx = cursorIdx
		// Try to include a few items before cursor
		w := parts[cursorIdx].width + arrowWidth
		for startIdx > 0 {
			prev := parts[startIdx-1].width + sepWidth
			if w+prev > width-arrowWidth {
				break
			}
			w += prev
			startIdx--
		}
	}

	// Build visible row
	hasLeft := startIdx > 0
	var row string
	if hasLeft {
		row = arrow
	}
	shown := 0
	hasRight := false
	for i := startIdx; i < len(parts); i++ {
		candidate := row
		if shown > 0 {
			candidate += sep
		}
		candidate += parts[i].text
		if lipgloss.Width(candidate)+arrowWidth > width && shown > 0 {
			hasRight = true
			break
		}
		row = candidate
		shown++
	}
	// Check if there are more items after what we rendered
	if startIdx+shown < len(parts) {
		hasRight = true
	}
	if hasRight {
		row += arrowR
	}

	barStyle := lipgloss.NewStyle().
		Background(colorSurface).
		Width(width).
		PaddingLeft(1)
	return barStyle.Render(row)
}
