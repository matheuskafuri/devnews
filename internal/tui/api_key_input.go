package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matheuskafuri/devnews/internal/ai"
	"github.com/matheuskafuri/devnews/internal/config"
)

func (a *App) handleAPIKeyInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.mode = modeNormal
		a.apiKeyInput.SetValue("")
		a.apiKeyInput.Blur()
		a.pendingSummary = false
		return a, nil
	case "enter":
		key := strings.TrimSpace(a.apiKeyInput.Value())
		if key == "" {
			return a, nil
		}
		a.apiKeyInput.Blur()
		return a, func() tea.Msg {
			if err := config.SaveAIKey("openai", key); err != nil {
				return feedErrMsg{err: err}
			}
			return apiKeySavedMsg{apiKey: key}
		}
	}

	var cmd tea.Cmd
	a.apiKeyInput, cmd = a.apiKeyInput.Update(msg)
	return a, cmd
}

func (a *App) handleAPIKeySaved(apiKey string) (tea.Model, tea.Cmd) {
	if a.cfg.AI == nil {
		a.cfg.AI = &config.AIConfig{Provider: "openai", Model: "gpt-4o-mini"}
	}
	a.cfg.AI.APIKey = apiKey
	s, err := ai.New(a.cfg.AI, apiKey)
	if err != nil {
		a.err = err
		a.mode = modeNormal
		return a, nil
	}
	a.summarizer = s
	a.mode = modeNormal
	a.apiKeyInput.SetValue("")

	if a.pendingSummary {
		a.pendingSummary = false
		if len(a.articles) > 0 && a.cursor < len(a.articles) {
			a.summaryLoading[a.articles[a.cursor].ID] = true
		}
		return a, a.fetchFullSummary()
	}
	return a, nil
}

func (a *App) renderAPIKeyOverlay() string {
	titleStyle := lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(colorText).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6272a4"))

	var b strings.Builder
	b.WriteString(titleStyle.Render("OpenAI API Key"))
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Enter your API key to enable AI summaries"))
	b.WriteString("\n\n")
	b.WriteString(a.apiKeyInput.View())
	b.WriteString("\n\n")
	b.WriteString(hintStyle.Render("enter save  esc cancel"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 3).
		Width(55)

	return boxStyle.Render(b.String())
}
