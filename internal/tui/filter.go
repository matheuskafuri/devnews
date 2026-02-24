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
	var parts []string

	// "All" tab
	if len(f.active) == 0 {
		parts = append(parts, tabActiveStyle.Render("All"))
	} else {
		parts = append(parts, tabInactiveStyle.Render("All"))
	}

	for i, s := range f.sources {
		style := tabInactiveStyle
		if f.active[s] {
			style = tabActiveStyle
		}
		label := s
		if f.filterMode && i == f.filterCursor {
			label = "[" + s + "]"
		}
		parts = append(parts, style.Render(label))
	}

	// Build row with · separators, stopping when we'd exceed width
	var row string
	for i, part := range parts {
		candidate := row
		if i > 0 {
			candidate += sep
		}
		candidate += part
		if lipgloss.Width(candidate) > width && row != "" {
			break
		}
		row = candidate
	}

	barStyle := lipgloss.NewStyle().
		Background(colorSurface).
		Width(width).
		PaddingLeft(1)
	return barStyle.Render(row)
}
