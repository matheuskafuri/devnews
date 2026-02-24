package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/matheuskafuri/devnews/internal/briefing"
	"github.com/matheuskafuri/devnews/internal/cache"
	"github.com/matheuskafuri/devnews/internal/signal"
)

func renderOpeningScreen(b *briefing.Briefing, height int) string {
	var lines []string

	title := briefingV2TitleStyle.Render(fmt.Sprintf("DevNews — %s", b.DateLabel))
	lines = append(lines, "", "  "+title, "")

	lines = append(lines, "  "+briefingV2MetaStyle.Render(fmt.Sprintf("Signal status: %s", b.Freshness)))
	lines = append(lines, "  "+briefingV2MetaStyle.Render(fmt.Sprintf("Posts scanned: %d", b.Scanned)))

	if b.Focus != "" {
		lines = append(lines, "  "+briefingV2MetaStyle.Render(fmt.Sprintf("%s articles: %d", b.Focus, b.Selected)))
	}

	lines = append(lines, "  "+briefingV2MetaStyle.Render(fmt.Sprintf("Selected for briefing: %d", b.Selected)))
	lines = append(lines, "")

	if len(b.Themes) > 0 {
		lines = append(lines, "  "+briefingV2BodyStyle.Render("Detected themes:"))
		for _, theme := range b.Themes {
			lines = append(lines, "  "+briefingV2BodyStyle.Render("  "+theme))
		}
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	contentLines := strings.Count(content, "\n") + 1
	topPad := (height - contentLines) / 3
	if topPad < 0 {
		topPad = 0
	}

	return strings.Repeat("\n", topPad) + content
}

func renderCardView(card briefing.Card, total int, width, height int, showBreakdown bool, sourceWeights signal.SourceWeights) string {
	// Card inner width
	cardWidth := width - 8
	if cardWidth < 30 {
		cardWidth = 30
	}

	// Build card body lines
	var body []string

	// Source · date line
	pubDate := card.Article.Published.Format("Jan 2")
	body = append(body, briefingV2MetaStyle.Render(card.Article.Source+" · "+pubDate))

	// Title
	body = append(body, briefingV2TitleStyle.Render(card.Article.Title))
	body = append(body, "")

	// Category · reading time · signal
	catStyle := categoryStyle(card.Article.Category)
	meta := catStyle.Render(card.Article.Category) +
		briefingV2MetaStyle.Render(fmt.Sprintf("  ·  %d min  ·  Signal %.1f", card.ReadingTime, card.Article.SignalScore))
	body = append(body, meta)

	// Why it matters
	if card.Article.WhyItMatters != "" {
		body = append(body, "")
		body = append(body, briefingV2WhyStyle.Render("Why it matters:"))
		wrapped := wrapText(card.Article.WhyItMatters, cardWidth-2)
		for _, line := range strings.Split(wrapped, "\n") {
			body = append(body, briefingV2WhyStyle.Render(line))
		}
	}

	cardContent := strings.Join(body, "\n")

	// Wrap in rounded border
	cardBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00E5FF")).
		Padding(0, 1).
		Width(cardWidth).
		Render(cardContent)

	// Counter above card
	counter := briefingV2MetaStyle.Render(fmt.Sprintf("%d/%d", card.Index, total))

	var lines []string
	lines = append(lines, "", "  "+counter)
	for _, l := range strings.Split(cardBox, "\n") {
		lines = append(lines, "  "+l)
	}

	content := strings.Join(lines, "\n")

	if showBreakdown {
		content += "\n\n" + renderSignalBreakdownOverlay(card.Article, sourceWeights)
	}

	contentLines := strings.Count(content, "\n") + 1
	topPad := (height - contentLines) / 3
	if topPad < 0 {
		topPad = 0
	}

	return strings.Repeat("\n", topPad) + content
}

func renderSignalBreakdownOverlay(a cache.Article, weights signal.SourceWeights) string {
	input := signal.Input{
		Title:       a.Title,
		Description: a.Description,
		Source:      a.Source,
		Published:   a.Published,
	}
	b := signal.ScoreWithBreakdown(input, weights)

	var lines []string
	lines = append(lines, "  "+briefingV2TitleStyle.Render("Signal Score Breakdown"))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Source weight:         %.2f", b.SourceWeight))
	lines = append(lines, fmt.Sprintf("  Depth score:           %.2f", b.Depth))
	lines = append(lines, fmt.Sprintf("  Recency score:         %.2f", b.Recency))
	lines = append(lines, fmt.Sprintf("  Keyword density score: %.2f", b.KeywordDensity))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  Final: %.1f", b.Final))

	styled := make([]string, len(lines))
	for i, l := range lines {
		styled[i] = briefingV2BodyStyle.Render(l)
	}
	return strings.Join(styled, "\n")
}
