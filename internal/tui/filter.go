package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const gridColumns = 3

type filterBar struct {
	sources    []string
	active     map[string]bool
	filterMode bool
	gridCursor int // 0 = "All", 1..len(sources) = individual sources
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

func (f *filterBar) toggleGrid() {
	if f.gridCursor == 0 {
		f.selectAll()
	} else {
		idx := f.gridCursor - 1
		if idx < len(f.sources) {
			f.toggle(f.sources[idx])
		}
	}
}

func (f *filterBar) selectAll() {
	f.active = make(map[string]bool)
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

func (f *filterBar) totalItems() int {
	return 1 + len(f.sources)
}

func rowsPerCol(total, cols int) int {
	return (total + cols - 1) / cols
}

// Grid navigation: items flow top-to-bottom, then left-to-right.
// Item i is at column i/rows, row i%rows.

func (f *filterBar) gridDown() {
	total := f.totalItems()
	rows := rowsPerCol(total, gridColumns)
	col := f.gridCursor / rows
	row := f.gridCursor % rows
	if row+1 < rows {
		next := col*rows + row + 1
		if next < total {
			f.gridCursor = next
		}
	}
}

func (f *filterBar) gridUp() {
	total := f.totalItems()
	rows := rowsPerCol(total, gridColumns)
	col := f.gridCursor / rows
	row := f.gridCursor % rows
	if row > 0 {
		f.gridCursor = col*rows + row - 1
	}
}

func (f *filterBar) gridLeft() {
	total := f.totalItems()
	rows := rowsPerCol(total, gridColumns)
	col := f.gridCursor / rows
	row := f.gridCursor % rows
	if col > 0 {
		f.gridCursor = (col-1)*rows + row
	}
}

func (f *filterBar) gridRight() {
	total := f.totalItems()
	rows := rowsPerCol(total, gridColumns)
	col := f.gridCursor / rows
	row := f.gridCursor % rows
	if col+1 < gridColumns {
		next := (col+1)*rows + row
		if next < total {
			f.gridCursor = next
		}
	}
}

// render produces the collapsed summary line shown in browse mode.
func (f *filterBar) render(width int) string {
	hintText := "f to filter"
	hint := lipgloss.NewStyle().Foreground(colorDim).Render(hintText)
	hintWidth := lipgloss.Width(hint)

	prefix := "Filter: "
	prefixWidth := lipgloss.Width(prefix)
	var label string
	if len(f.active) == 0 {
		label = "All sources"
	} else {
		names := f.activeSources()
		label = strings.Join(names, ", ")
		maxLabelWidth := width - prefixWidth - hintWidth - 4 // 4 = padding + gaps
		if maxLabelWidth > 0 && lipgloss.Width(label) > maxLabelWidth {
			// Truncate: show as many names as fit + "(+N more)"
			label = ""
			shown := 0
			for i, name := range names {
				candidate := label
				if shown > 0 {
					candidate += ", "
				}
				remaining := len(names) - i - 1
				more := fmt.Sprintf(" (+%d more)", remaining)
				if lipgloss.Width(candidate+name+more) > maxLabelWidth {
					if shown == 0 {
						label = fmt.Sprintf("(%d sources)", len(names))
					} else {
						label = candidate + fmt.Sprintf("(+%d more)", remaining)
					}
					break
				}
				candidate += name
				label = candidate
				shown++
			}
		}
	}

	left := lipgloss.NewStyle().Foreground(colorText).Render(prefix) +
		lipgloss.NewStyle().Foreground(colorAccent).Render(label)
	leftWidth := lipgloss.Width(left)
	gap := width - leftWidth - hintWidth - 2 // 2 = PaddingLeft(1) + 1 extra
	if gap < 1 {
		gap = 1
	}

	row := left + strings.Repeat(" ", gap) + hint

	barStyle := lipgloss.NewStyle().
		Background(colorSurface).
		Width(width).
		PaddingLeft(1)
	return barStyle.Render(row)
}

// Styles used inside renderOverlay, allocated once.
var (
	overlayActiveNameStyle   = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	overlayInactiveNameStyle = lipgloss.NewStyle().Foreground(colorText)
	overlayCursorStyle       = lipgloss.NewStyle().Foreground(colorAccent)
	overlayCheckOnStyle      = lipgloss.NewStyle().Foreground(colorAccent)
	overlayCheckOffStyle     = lipgloss.NewStyle().Foreground(colorDim)
	overlayHelpStyle         = lipgloss.NewStyle().Foreground(colorDim)
)

// renderOverlay produces the full filter overlay grid panel.
func (f *filterBar) renderOverlay() string {
	total := f.totalItems()
	rows := rowsPerCol(total, gridColumns)

	// Build item labels: ["All", source0, source1, ...]
	items := make([]string, total)
	items[0] = "All"
	for i, s := range f.sources {
		items[i+1] = s
	}

	// Determine if each item is active
	isActive := make([]bool, total)
	isActive[0] = len(f.active) == 0
	for i, s := range f.sources {
		isActive[i+1] = f.active[s]
	}

	// Find max item name width for uniform columns
	maxNameWidth := 0
	for _, name := range items {
		w := lipgloss.Width(name)
		if w > maxNameWidth {
			maxNameWidth = w
		}
	}
	// Each cell: "> [x] Name" = 2 (cursor) + 4 ([x] ) + name
	cellWidth := 2 + 4 + maxNameWidth

	// Build columns
	columns := make([]string, gridColumns)
	for col := 0; col < gridColumns; col++ {
		var lines []string
		for row := 0; row < rows; row++ {
			idx := col*rows + row
			if idx >= total {
				lines = append(lines, strings.Repeat(" ", cellWidth))
				continue
			}

			// Cursor indicator
			cursor := "  "
			if f.filterMode && idx == f.gridCursor {
				cursor = overlayCursorStyle.Render("> ")
			}

			// Checkbox
			var checkbox string
			if isActive[idx] {
				checkbox = overlayCheckOnStyle.Render("[x] ")
			} else {
				checkbox = overlayCheckOffStyle.Render("[ ] ")
			}

			// Name
			name := items[idx]
			var nameStyled string
			if isActive[idx] {
				nameStyled = overlayActiveNameStyle.Render(name)
			} else {
				nameStyled = overlayInactiveNameStyle.Render(name)
			}

			// Pad name to fixed visual width
			pad := maxNameWidth - lipgloss.Width(name)
			if pad < 0 {
				pad = 0
			}
			line := cursor + checkbox + nameStyled + strings.Repeat(" ", pad)
			lines = append(lines, line)
		}
		columns[col] = strings.Join(lines, "\n")
	}

	grid := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	// Title
	title := briefingV2TitleStyle.Render("Filter Sources")

	// Help line
	help := overlayHelpStyle.Render("↑↓←→ navigate  space toggle  a all  esc close")

	// Compose panel content
	content := title + "\n\n" + grid + "\n\n" + help

	// Apply overlay style and center
	return filterOverlayStyle.Render(content)
}

// overlayCenter composites the overlay panel on top of the background view.
// It splits both into lines and replaces the center rows of bg with overlay rows.
func overlayCenter(bg, overlay string, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	// Pad bg to full height
	for len(bgLines) < height {
		bgLines = append(bgLines, "")
	}

	panel := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, overlay)
	panelLines := strings.Split(panel, "\n")

	// Find the vertical range where the overlay has content (non-empty lines)
	top, bottom := -1, -1
	for i, line := range panelLines {
		if strings.TrimSpace(line) != "" {
			if top == -1 {
				top = i
			}
			bottom = i
		}
	}

	if top == -1 {
		return bg // no overlay content
	}

	// Replace bg lines with overlay lines in the overlay region
	result := make([]string, len(bgLines))
	copy(result, bgLines)
	for i := top; i <= bottom && i < len(result) && i < len(panelLines); i++ {
		result[i] = panelLines[i]
	}

	return strings.Join(result, "\n")
}
