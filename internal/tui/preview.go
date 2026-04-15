package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/matheuskafuri/devnews/internal/cache"
)

func renderPreview(article *cache.Article, width, height, scroll int, loadingSummary bool) string {
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

	rule := lipgloss.NewStyle().Foreground(colorSubtle).Render(strings.Repeat("─", contentWidth))

	var parts []string
	parts = append(parts, title, source, rule)

	// AI summary
	if article.Summary != "" {
		summary := previewSummaryStyle.Width(contentWidth).Render("░ " + article.Summary)
		parts = append(parts, summary)
		if article.Tags != "" {
			tags := previewTagsStyle.Render(article.Tags)
			parts = append(parts, tags)
		}
		parts = append(parts, "")
	}

	// Full AI article summary
	if article.FullSummary != "" {
		label := fullSummaryLabelStyle.Width(contentWidth).Render("AI Summary")
		body := fullSummaryStyle.Width(contentWidth).Render(wrapText(article.FullSummary, contentWidth))
		parts = append(parts, "", label, body)
	}

	// Loading indicator
	if loadingSummary && article.FullSummary == "" {
		loading := fullSummaryLabelStyle.Width(contentWidth).Render("Generating AI summary...")
		parts = append(parts, "", loading)
	}

	desc := article.Description
	if desc == "" {
		desc = "(No description available)"
	}

	body := previewBodyStyle.Width(contentWidth).Render(wrapText(desc, contentWidth))
	link := previewLinkStyle.Width(contentWidth).Render("Read more: " + article.Link)

	hint := lipgloss.NewStyle().Foreground(colorDim).MarginTop(1).Render("Press Enter to open in browser")
	parts = append(parts, body, "", link, hint)
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

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
