package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) handleRequestSourceKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.mode = modeHome
		a.sourceNameInput.Blur()
		a.sourceURLInput.Blur()
		return a, nil
	case "tab", "shift+tab":
		if a.sourceFormFocus == 0 {
			a.sourceFormFocus = 1
			a.sourceNameInput.Blur()
			a.sourceURLInput.Focus()
		} else {
			a.sourceFormFocus = 0
			a.sourceURLInput.Blur()
			a.sourceNameInput.Focus()
		}
		return a, textinput.Blink
	case "enter":
		name := strings.TrimSpace(a.sourceNameInput.Value())
		url := strings.TrimSpace(a.sourceURLInput.Value())
		if name == "" || url == "" {
			a.statusMessage = "Both name and URL are required"
			return a, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
				return clearStatusMsg{}
			})
		}
		a.sourceNameInput.Blur()
		a.sourceURLInput.Blur()
		return a, submitSourceRequest(name, url)
	}

	var cmd tea.Cmd
	if a.sourceFormFocus == 0 {
		a.sourceNameInput, cmd = a.sourceNameInput.Update(msg)
	} else {
		a.sourceURLInput, cmd = a.sourceURLInput.Update(msg)
	}
	return a, cmd
}

func submitSourceRequest(name, url string) tea.Cmd {
	return func() tea.Msg {
		title := fmt.Sprintf("[Source Request] %s", name)
		body := fmt.Sprintf("**Source Name:** %s\n**Feed URL:** %s\n\nRequested via devnews TUI.", name, url)

		if _, lookErr := exec.LookPath("gh"); lookErr != nil {
			return sourceRequestResultMsg{err: fmt.Errorf("gh CLI not found — install from https://cli.github.com")}
		}

		cmd := exec.Command("gh", "issue", "create",
			"--repo", "matheuskafuri/devnews",
			"--title", title,
			"--body", body,
			"--label", "source-request",
		)
		output, err := cmd.CombinedOutput()
		if err != nil && strings.Contains(string(output), "label") {
			// Retry without label if it doesn't exist
			cmd = exec.Command("gh", "issue", "create",
				"--repo", "matheuskafuri/devnews",
				"--title", title,
				"--body", body,
			)
			output, err = cmd.CombinedOutput()
		}
		if err != nil {
			msg := strings.TrimSpace(string(output))
			if msg == "" {
				msg = err.Error()
			}
			return sourceRequestResultMsg{err: fmt.Errorf("%s", msg)}
		}
		return sourceRequestResultMsg{}
	}
}

func (a *App) renderRequestSourceOverlay() string {
	var b strings.Builder
	b.WriteString(overlayTitleStyle.Render("Request a Source"))
	b.WriteString("\n\n")
	b.WriteString(overlayLabelStyle.Render("Name"))
	b.WriteString("\n")
	b.WriteString(a.sourceNameInput.View())
	b.WriteString("\n\n")
	b.WriteString(overlayLabelStyle.Render("URL"))
	b.WriteString("\n")
	b.WriteString(a.sourceURLInput.View())
	b.WriteString("\n\n")
	b.WriteString(overlayHintStyle.Render("tab switch  enter submit  esc cancel"))

	return overlayBoxStyle(50).Render(b.String())
}
