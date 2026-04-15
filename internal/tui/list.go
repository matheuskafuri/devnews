package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/matheuskafuri/devnews/internal/cache"
)

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	default:
		return t.Format("Jan 2")
	}
}

func timeColor(published time.Time) lipgloss.Style {
	age := time.Since(published)
	switch {
	case age < 6*time.Hour:
		return lipgloss.NewStyle().Foreground(colorAccent) // bright/fresh
	case age < 24*time.Hour:
		return itemTimeStyle // normal
	default:
		return lipgloss.NewStyle().Foreground(colorDim) // old/faded
	}
}

func renderListItem(a cache.Article, selected bool, width int) string {
	if width < 10 {
		width = 30
	}

	// Line 1: indicator + title + right-aligned markers (AI tag + time)
	timeStr := relativeTime(a.Published)
	timeStyle := timeColor(a.Published)

	// Right side: optional AI marker + time
	var rightParts []string
	if a.FullSummary != "" {
		rightParts = append(rightParts, itemAIMarkerStyle.Render("AI"))
	}
	rightParts = append(rightParts, timeStyle.Render(timeStr))
	right := strings.Join(rightParts, " ")
	rightWidth := lipgloss.Width(right)

	// Left side: indicator + title
	var indicator string
	var titleStyle lipgloss.Style
	if selected {
		indicator = itemSelectedStyle.Render("▸ ")
		titleStyle = itemSelectedStyle
	} else if a.Read {
		indicator = itemReadStyle.Render("○ ")
		titleStyle = itemReadStyle
	} else {
		indicator = itemUnreadStyle.Render("● ")
		titleStyle = itemTitleStyle
	}

	maxTitle := width - 4 - rightWidth // 2 for indicator, 2 for gap
	if maxTitle < 10 {
		maxTitle = 10
	}
	titleStr := titleStyle.Render(truncateStr(a.Title, maxTitle))

	gap := width - lipgloss.Width(indicator) - lipgloss.Width(titleStr) - rightWidth
	if gap < 1 {
		gap = 1
	}
	line1 := indicator + titleStr + strings.Repeat(" ", gap) + right

	// Line 2: category badge (colored)
	line2 := "    "
	if a.Category != "" {
		line2 += categoryStyle(a.Category).Render(a.Category)
	} else {
		line2 += itemSourceStyle.Render(a.Source)
	}

	return line1 + "\n" + line2
}

func truncateStr(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 3 {
		return string(runes[:n])
	}
	return string(runes[:n-3]) + "..."
}

func renderList(articles []cache.Article, cursor int, height int, width int) string {
	if len(articles) == 0 {
		return lipglossCenter("No articles found", width, height)
	}

	// Each item is 2 lines + 1 blank line = 3 lines
	itemHeight := 3
	visible := height / itemHeight
	if visible < 1 {
		visible = 1
	}

	// Calculate scroll offset
	start := 0
	if cursor >= visible {
		start = cursor - visible + 1
	}
	end := start + visible
	if end > len(articles) {
		end = len(articles)
		start = end - visible
		if start < 0 {
			start = 0
		}
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		b.WriteString(renderListItem(articles[i], i == cursor, width))
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func lipglossCenter(s string, width, height int) string {
	return strings.Repeat("\n", height/3) + strings.Repeat(" ", (width-len(s))/2) + s
}
