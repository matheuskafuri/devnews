package tui

import (
	"fmt"
	"strings"
	"time"

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

func renderListItem(a cache.Article, selected bool, width int) string {
	if width < 10 {
		width = 30
	}

	var title string
	if selected {
		title = itemSelectedStyle.Render("> " + truncateStr(a.Title, width-4))
	} else {
		title = itemTitleStyle.Render("  " + truncateStr(a.Title, width-4))
	}

	meta := "  " + itemSourceStyle.Render(a.Source) + " " + itemTimeStyle.Render("· "+relativeTime(a.Published))

	return title + "\n" + meta
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
