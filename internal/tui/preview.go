package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/matheuskafuri/devnews/internal/cache"
)

func renderPreview(article *cache.Article, width, height, scroll int) string {
	if article == nil {
		return lipglossCenter("Select an article", width, height)
	}

	contentWidth := width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	title := previewTitleStyle.Width(contentWidth).Render(article.Title)
	source := previewSourceStyle.Render(
		fmt.Sprintf("%s · %s", article.Source, article.Published.Format("Jan 2, 2006")),
	)

	desc := article.Description
	if desc == "" {
		desc = "(No description available)"
	}

	body := previewBodyStyle.Width(contentWidth).Render(wrapText(desc, contentWidth))
	link := previewLinkStyle.Width(contentWidth).Render("Read more: " + article.Link)

	content := lipgloss.JoinVertical(lipgloss.Left, title, source, "", body, "", link)

	// Apply scroll offset
	lines := strings.Split(content, "\n")
	if scroll > 0 && scroll < len(lines) {
		lines = lines[scroll:]
	}

	// Pad to fill height
	if len(lines) < height {
		lines = append(lines, make([]string, height-len(lines))...)
	} else if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}
