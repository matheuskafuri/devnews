package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matheuskafuri/devnews/internal/browser"
	"github.com/matheuskafuri/devnews/internal/cache"
	"github.com/matheuskafuri/devnews/internal/config"
	"github.com/matheuskafuri/devnews/internal/feed"
)

type focusPane int

const (
	focusList focusPane = iota
	focusPreview
)

type mode int

const (
	modeNormal mode = iota
	modeSearch
	modeFilter
	modeHelp
)

type App struct {
	cfg      *config.Config
	db       *cache.Cache
	articles []cache.Article
	cursor   int
	focus    focusPane
	mode     mode

	width  int
	height int

	// Sub-components
	searchInput textinput.Model
	spinner     spinner.Model
	filterBar   filterBar

	// State
	refreshing    bool
	since         time.Time
	previewScroll int
	currentDate   string
	err           error
}

func NewApp(cfg *config.Config, db *cache.Cache, since time.Time) *App {
	ti := textinput.New()
	ti.Placeholder = "Search articles..."
	ti.Prompt = searchPromptStyle.Render("/ ")
	ti.CharLimit = 100

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = spinnerStyle

	return &App{
		cfg:         cfg,
		db:          db,
		since:       since,
		filterBar:   newFilterBar(cfg.SourceNames()),
		searchInput: ti,
		spinner:     sp,
		currentDate: time.Now().Format("Jan 2"),
	}
}

func (a *App) Init() tea.Cmd {
	return a.loadArticlesCmd()
}

// loadArticlesCmd captures current query state into the closure to avoid races.
func (a *App) loadArticlesCmd() tea.Cmd {
	opts := cache.QueryOpts{
		Since:   a.since,
		Sources: a.filterBar.activeSources(),
		Search:  a.searchInput.Value(),
	}
	db := a.db
	return func() tea.Msg {
		articles, err := db.GetArticles(opts)
		if err != nil {
			return feedErrMsg{err: err}
		}
		return feedsLoadedMsg{articles: articles}
	}
}

func (a *App) doRefresh() tea.Cmd {
	cfg := a.cfg
	db := a.db
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result := feed.FetchAll(ctx, cfg.EnabledSources())

		if err := db.UpsertArticles(result.Articles); err != nil {
			return refreshDoneMsg{errs: append(result.Errors, err)}
		}
		db.SetLastRefresh()

		return refreshDoneMsg{count: len(result.Articles), errs: result.Errors}
	}
}

func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		err := browser.Open(url)
		if err != nil {
			return feedErrMsg{err: err}
		}
		return nil
	}
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		// Clear sticky error on any keypress
		a.err = nil
		return a.handleKey(msg)

	case feedsLoadedMsg:
		a.articles = msg.articles
		if a.cursor >= len(a.articles) {
			a.cursor = max(0, len(a.articles)-1)
		}
		return a, nil

	case feedErrMsg:
		a.err = msg.err
		return a, nil

	case refreshDoneMsg:
		a.refreshing = false
		return a, a.loadArticlesCmd()

	case spinner.TickMsg:
		if a.refreshing {
			var cmd tea.Cmd
			a.spinner, cmd = a.spinner.Update(msg)
			return a, cmd
		}
		return a, nil
	}

	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c":
		return a, tea.Quit
	}

	// Mode-specific handling
	switch a.mode {
	case modeSearch:
		return a.handleSearchKey(msg)
	case modeFilter:
		return a.handleFilterKey(msg)
	case modeHelp:
		if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
			a.mode = modeNormal
		}
		return a, nil
	}

	// Normal mode
	switch msg.String() {
	case "q":
		return a, tea.Quit
	case "j", "down":
		if a.focus == focusList && a.cursor < len(a.articles)-1 {
			a.cursor++
			a.previewScroll = 0
		} else if a.focus == focusPreview {
			a.previewScroll++
		}
		return a, nil
	case "k", "up":
		if a.focus == focusList && a.cursor > 0 {
			a.cursor--
			a.previewScroll = 0
		} else if a.focus == focusPreview && a.previewScroll > 0 {
			a.previewScroll--
		}
		return a, nil
	case "tab":
		if a.focus == focusList {
			a.focus = focusPreview
		} else {
			a.focus = focusList
		}
		return a, nil
	case "o", "enter":
		if len(a.articles) > 0 && a.cursor < len(a.articles) {
			return a, openBrowserCmd(a.articles[a.cursor].Link)
		}
		return a, nil
	case "/":
		a.mode = modeSearch
		a.searchInput.Focus()
		return a, textinput.Blink
	case "f":
		a.mode = modeFilter
		a.filterBar.filterMode = true
		return a, nil
	case "r":
		if !a.refreshing {
			a.refreshing = true
			return a, tea.Batch(a.doRefresh(), a.spinner.Tick)
		}
		return a, nil
	case "?":
		a.mode = modeHelp
		return a, nil
	}

	return a, nil
}

func (a *App) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.mode = modeNormal
		a.searchInput.SetValue("")
		a.searchInput.Blur()
		return a, a.loadArticlesCmd()
	case "enter":
		a.mode = modeNormal
		a.searchInput.Blur()
		return a, a.loadArticlesCmd()
	}

	var cmd tea.Cmd
	a.searchInput, cmd = a.searchInput.Update(msg)
	// Only re-query on actual value changes, not cursor moves etc.
	return a, cmd
}

func (a *App) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "f":
		a.mode = modeNormal
		a.filterBar.filterMode = false
		return a, nil
	case "left", "h":
		if a.filterBar.filterCursor > 0 {
			a.filterBar.filterCursor--
		}
		return a, nil
	case "right", "l":
		if a.filterBar.filterCursor < len(a.filterBar.sources)-1 {
			a.filterBar.filterCursor++
		}
		return a, nil
	case " ", "enter":
		a.filterBar.toggleCurrent()
		a.cursor = 0
		return a, a.loadArticlesCmd()
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx := int(msg.String()[0] - '1')
		if idx < len(a.filterBar.sources) {
			a.filterBar.toggle(a.filterBar.sources[idx])
			a.cursor = 0
			return a, a.loadArticlesCmd()
		}
		return a, nil
	}
	return a, nil
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	if a.mode == modeHelp {
		return a.renderHelp()
	}

	// Layout calculations
	headerHeight := 1
	filterHeight := 1
	statusHeight := 1
	contentHeight := a.height - headerHeight - filterHeight - statusHeight - 4 // borders

	listWidth := int(float64(a.width) * 0.35)
	previewWidth := a.width - listWidth - 1 // gap

	if contentHeight < 3 {
		contentHeight = 3
	}

	// Header
	headerLeft := headerStyle.Render("devnews — Engineering Blog Aggregator")
	headerRight := headerDateStyle.Render(a.currentDate)
	headerGap := a.width - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight)
	if headerGap < 0 {
		headerGap = 0
	}
	header := headerLeft + fmt.Sprintf("%*s", headerGap, "") + headerRight

	// Filter bar
	filter := a.filterBar.render(a.width)

	// Search bar (replaces filter when searching)
	if a.mode == modeSearch {
		filter = a.searchInput.View()
	}

	// List pane
	innerListW := listWidth - 4 // border + padding
	listContent := renderList(a.articles, a.cursor, contentHeight, innerListW)

	var listPane string
	if a.focus == focusList {
		listPane = listPaneActiveStyle.Width(listWidth - 2).Height(contentHeight).Render(listContent)
	} else {
		listPane = listPaneStyle.Width(listWidth - 2).Height(contentHeight).Render(listContent)
	}

	// Preview pane
	var selected *cache.Article
	if len(a.articles) > 0 && a.cursor < len(a.articles) {
		selected = &a.articles[a.cursor]
	}
	innerPreviewW := previewWidth - 4
	previewContent := renderPreview(selected, innerPreviewW, contentHeight, a.previewScroll)

	var previewPane string
	if a.focus == focusPreview {
		previewPane = previewPaneActiveStyle.Width(previewWidth - 2).Height(contentHeight).Render(previewContent)
	} else {
		previewPane = previewPaneStyle.Width(previewWidth - 2).Height(contentHeight).Render(previewContent)
	}

	// Join panes
	content := lipgloss.JoinHorizontal(lipgloss.Top, listPane, previewPane)

	// Status bar
	status := renderStatusBar(
		len(a.articles),
		a.filterBar.activeLabel(),
		a.width,
		a.mode == modeSearch,
		a.refreshing,
	)

	if a.refreshing {
		status = a.spinner.View() + " " + status
	}

	// Error display
	if a.err != nil {
		status = lipgloss.NewStyle().Foreground(colorAccent).Render(a.err.Error())
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, filter, content, status)
}

func (a *App) renderHelp() string {
	help := `
  devnews — Keyboard Shortcuts

  Navigation
    j/k, ↑/↓     Navigate article list
    tab           Switch focus between list and preview

  Actions
    o, enter      Open article in browser
    r             Refresh feeds
    /             Search articles
    f             Toggle source filter mode

  Filter Mode
    ←/→, h/l     Move between sources
    space/enter   Toggle source
    1-9           Toggle source by number
    esc, f        Exit filter mode

  General
    ?             Toggle this help
    q, ctrl+c    Quit
`
	style := lipgloss.NewStyle().
		Width(a.width).
		Height(a.height).
		Align(lipgloss.Center, lipgloss.Center).
		Foreground(colorSecondary)

	return style.Render(help)
}

// Run starts the TUI application.
func Run(cfg *config.Config, db *cache.Cache, since time.Time) error {
	app := NewApp(cfg, db, since)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
