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

	rule := previewRuleStyle.Render(strings.Repeat("─", contentWidth))

	// Header section
	title := previewTitleStyle.Width(contentWidth).Render(article.Title)
	source := previewSourceStyle.Render(
		fmt.Sprintf("%s · %s", article.Source, article.Published.Format("Jan 2, 2006")),
	)

	var parts []string
	parts = append(parts, title, source)

	// Category (colored)
	if article.Category != "" {
		cat := categoryStyle(article.Category).Render(article.Category)
		parts = append(parts, cat)
	}

	parts = append(parts, rule)

	// AI Summary section
	if article.FullSummary != "" {
		sectionHeader := fullSummaryLabelStyle.Width(contentWidth).Render("░ AI Summary")
		body := fullSummaryStyle.Width(contentWidth).Render(wrapText(article.FullSummary, contentWidth))
		parts = append(parts, "", sectionHeader, body, "", rule)
	} else if article.Summary != "" {
		sectionHeader := fullSummaryLabelStyle.Width(contentWidth).Render("░ Summary")
		summary := fullSummaryStyle.Width(contentWidth).Render(wrapText(article.Summary, contentWidth))
		parts = append(parts, "", sectionHeader, summary)
		if article.Tags != "" {
			tags := previewTagsStyle.Render(article.Tags)
			parts = append(parts, tags)
		}
		parts = append(parts, "", rule)
	}

	// Loading indicator
	if loadingSummary && article.FullSummary == "" {
		loading := fullSummaryLabelStyle.Width(contentWidth).Render("░░░▒▒▒▓▓▓ Generating summary...")
		parts = append(parts, "", loading, "")
	}

	// Description section
	desc := article.Description
	if desc == "" {
		desc = "(No description available)"
	}
	descHeader := previewSourceStyle.Render("Description")
	body := previewBodyStyle.Width(contentWidth).Render(wrapText(desc, contentWidth))
	parts = append(parts, "", descHeader, body)

	// Link
	link := previewLinkStyle.Width(contentWidth).Render("Read more: " + article.Link)
	parts = append(parts, "", link)

	// Bottom rule + hints
	parts = append(parts, rule)
	hint := previewHintStyle.Render("S summarize  o open  v layout")
	parts = append(parts, hint)

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Apply scroll offset
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	if scroll > 0 && scroll < len(lines) {
		lines = lines[scroll:]
	}

	// Pad or truncate to fill height
	if len(lines) < height {
		lines = append(lines, make([]string, height-len(lines))...)
	} else if len(lines) > height {
		lines = lines[:height]
	}

	// Scroll indicators
	if scroll > 0 {
		// Show ▲ indicator at top-right
		if len(lines) > 0 {
			indicator := previewRuleStyle.Render("▲ more")
			pad := contentWidth - lipgloss.Width(indicator)
			if pad < 0 {
				pad = 0
			}
			lines[0] = strings.Repeat(" ", pad) + indicator
		}
	}
	if totalLines > scroll+height {
		// Show ▼ indicator at bottom-right
		if len(lines) > 0 {
			indicator := previewRuleStyle.Render("▼ more")
			pad := contentWidth - lipgloss.Width(indicator)
			if pad < 0 {
				pad = 0
			}
			lines[len(lines)-1] = strings.Repeat(" ", pad) + indicator
		}
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
