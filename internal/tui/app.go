package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/matheuskafuri/devnews/internal/ai"
	"github.com/matheuskafuri/devnews/internal/briefing"
	"github.com/matheuskafuri/devnews/internal/browser"
	"github.com/matheuskafuri/devnews/internal/cache"
	"github.com/matheuskafuri/devnews/internal/config"
	"github.com/matheuskafuri/devnews/internal/feed"
	"github.com/matheuskafuri/devnews/internal/signal"
)

type focusPane int

const (
	focusList focusPane = iota
	focusPreview
)

type mode int

const (
	modeHome mode = iota
	modeNormal
	modeSearch
	modeFilter
	modeHelp
	modeBriefingOpening
	modeBriefingCard
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

	// AI
	summarizer ai.Summarizer

	// State
	refreshing         bool
	since              time.Time
	previewScroll      int
	currentDate        string
	streak             int
	err                error
	briefingV2          *briefing.Briefing
	cardCursor          int
	showSignalBreakdown bool
	sourceWeights      signal.SourceWeights
}

// RunOpts holds all parameters for launching the TUI.
type RunOpts struct {
	Cfg           *config.Config
	DB            *cache.Cache
	Since         time.Time
	Streak        int
	Summarizer    ai.Summarizer
	BrowseMode    bool
	BriefingV2    *briefing.Briefing
	SourceWeights signal.SourceWeights
}

func NewApp(opts RunOpts) *App {
	ti := textinput.New()
	ti.Placeholder = "Search articles..."
	ti.Prompt = searchPromptStyle.Render("/ ")
	ti.CharLimit = 100

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = spinnerStyle

	startMode := modeHome
	if opts.BrowseMode {
		startMode = modeNormal
	}

	return &App{
		cfg:           opts.Cfg,
		db:            opts.DB,
		since:         opts.Since,
		streak:        opts.Streak,
		summarizer:    opts.Summarizer,
		filterBar:     newFilterBar(opts.Cfg.SourceNames()),
		searchInput:   ti,
		spinner:       sp,
		currentDate:   time.Now().Format("Jan 2"),
		mode:          startMode,
		briefingV2:    opts.BriefingV2,
		sourceWeights: opts.SourceWeights,
	}
}

func (a *App) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Only load articles immediately if starting in browse mode
	if a.mode == modeNormal {
		cmds = append(cmds, a.loadArticlesCmd())
	}

	// Async AI enrichment for V2 briefing
	if a.summarizer != nil && a.briefingV2 != nil {
		cmds = append(cmds, a.fetchWhyItMatters()...)
		cmds = append(cmds, a.fetchThemes())
	}

	return tea.Batch(cmds...)
}

func (a *App) fetchWhyItMatters() []tea.Cmd {
	var cmds []tea.Cmd
	s := a.summarizer
	db := a.db
	for i, card := range a.briefingV2.Cards {
		if card.Article.WhyItMatters != "" && card.Article.WhyItMatters != briefing.DescriptionExcerpt(card.Article.Description) {
			continue // already has AI-generated text
		}
		idx := i
		art := card.Article
		cmds = append(cmds, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			text, err := s.WhyItMatters(ctx, art.Title, art.Description)
			if err != nil || text == "" {
				return nil
			}
			db.UpdateArticleWhyItMatters(art.ID, text)
			return whyItMattersMsg{cardIndex: idx, articleID: art.ID, text: text}
		})
	}
	return cmds
}

func (a *App) fetchThemes() tea.Cmd {
	s := a.summarizer
	cards := a.briefingV2.Cards
	summaries := make([]ai.ArticleSummary, len(cards))
	for i, c := range cards {
		summaries[i] = ai.ArticleSummary{Title: c.Article.Title, Category: c.Article.Category}
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		themes, err := s.Themes(ctx, summaries)
		if err != nil || len(themes) == 0 {
			return nil
		}
		return themesMsg{themes: themes}
	}
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
		return a, a.maybeFetchSummary()

	case feedErrMsg:
		a.err = msg.err
		return a, nil

	case refreshDoneMsg:
		a.refreshing = false
		return a, a.loadArticlesCmd()

	case summaryLoadedMsg:
		// Update the article in our local slice
		tags := strings.Join(msg.result.Tags, ", ")
		for i := range a.articles {
			if a.articles[i].ID == msg.articleID {
				a.articles[i].Summary = msg.result.Summary
				a.articles[i].Tags = tags
				break
			}
		}
		// Persist to cache asynchronously
		db := a.db
		id := msg.articleID
		summary := msg.result.Summary
		return a, func() tea.Msg {
			db.UpdateArticleSummary(id, summary, tags)
			return nil
		}

	case whyItMattersMsg:
		if a.briefingV2 != nil && msg.cardIndex < len(a.briefingV2.Cards) {
			a.briefingV2.Cards[msg.cardIndex].Article.WhyItMatters = msg.text
		}
		return a, nil

	case themesMsg:
		if a.briefingV2 != nil {
			a.briefingV2.Themes = msg.themes
		}
		return a, nil

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
	case modeHome:
		return a.handleHomeKey(msg)
	case modeBriefingOpening:
		return a.handleBriefingOpeningKey(msg)
	case modeBriefingCard:
		return a.handleBriefingCardKey(msg)
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
			return a, a.maybeFetchSummary()
		} else if a.focus == focusPreview {
			a.previewScroll++
		}
		return a, nil
	case "k", "up":
		if a.focus == focusList && a.cursor > 0 {
			a.cursor--
			a.previewScroll = 0
			return a, a.maybeFetchSummary()
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
	case "h":
		a.mode = modeHome
		return a, nil
	case "?":
		a.mode = modeHelp
		return a, nil
	}

	return a, nil
}

func (a *App) handleHomeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b", "1":
		if a.briefingV2 != nil && len(a.briefingV2.Cards) > 0 {
			a.mode = modeBriefingOpening
			return a, nil
		}
		a.mode = modeNormal
		return a, a.loadArticlesCmd()
	case "e", "2":
		a.mode = modeNormal
		return a, a.loadArticlesCmd()
	case "q":
		return a, tea.Quit
	}
	return a, nil
}

func (a *App) handleBriefingOpeningKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		a.mode = modeBriefingCard
		a.cardCursor = 0
		return a, nil
	case "q":
		return a, tea.Quit
	case "e":
		a.mode = modeNormal
		return a, a.loadArticlesCmd()
	case "h":
		a.mode = modeHome
		return a, nil
	}
	return a, nil
}

func (a *App) handleBriefingCardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n", "j", "right":
		if a.briefingV2 != nil && a.cardCursor < len(a.briefingV2.Cards)-1 {
			a.cardCursor++
			a.showSignalBreakdown = false
		}
		return a, nil
	case "p", "k", "left":
		if a.cardCursor > 0 {
			a.cardCursor--
			a.showSignalBreakdown = false
		}
		return a, nil
	case "o", "enter":
		if a.briefingV2 != nil && a.cardCursor < len(a.briefingV2.Cards) {
			return a, openBrowserCmd(a.briefingV2.Cards[a.cardCursor].Article.Link)
		}
		return a, nil
	case "i":
		a.showSignalBreakdown = !a.showSignalBreakdown
		return a, nil
	case "e":
		a.mode = modeNormal
		return a, a.loadArticlesCmd()
	case "h":
		a.mode = modeHome
		return a, nil
	case "q":
		return a, tea.Quit
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

func (a *App) withBottomBar(content string, hints string) string {
	bar := renderBottomBar(a.streak, hints, a.width)
	lines := strings.Split(content, "\n")
	for len(lines) < a.height-1 {
		lines = append(lines, "")
	}
	if len(lines) >= a.height {
		lines = lines[:a.height-1]
	}
	lines = append(lines, bar)
	return strings.Join(lines, "\n")
}

func (a *App) View() string {
	if a.width == 0 {
		return lipgloss.NewStyle().Foreground(colorAccent).Render("  devnews")
	}

	if a.mode == modeHome {
		hasBriefing := a.briefingV2 != nil && len(a.briefingV2.Cards) > 0
		return a.withBottomBar(renderHomeScreen(a.width, a.height, hasBriefing), "b briefing  e browse  q quit")
	}

	if a.mode == modeBriefingOpening && a.briefingV2 != nil {
		return a.withBottomBar(renderOpeningScreen(a.briefingV2, a.height), "enter start  e browse  h home  q quit")
	}

	if a.mode == modeBriefingCard && a.briefingV2 != nil && a.cardCursor < len(a.briefingV2.Cards) {
		return a.withBottomBar(
			renderCardView(a.briefingV2.Cards[a.cardCursor], len(a.briefingV2.Cards), a.width, a.height, a.showSignalBreakdown, a.sourceWeights),
			"n next  p prev  o open  i info  e browse  h home  q quit",
		)
	}

	if a.mode == modeHelp {
		return a.withBottomBar(a.renderHelp(), "? close  h home  q quit")
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
	headerLeft := headerStyle.Render("devnews")
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
		a.streak,
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

func (a *App) maybeFetchSummary() tea.Cmd {
	if a.summarizer == nil {
		return nil
	}
	if len(a.articles) == 0 || a.cursor >= len(a.articles) {
		return nil
	}
	article := a.articles[a.cursor]
	if article.Summary != "" {
		return nil // already cached
	}
	s := a.summarizer
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		result, err := s.Summarize(ctx, article.Title, article.Description)
		if err != nil {
			return nil // non-fatal
		}
		return summaryLoadedMsg{articleID: article.ID, result: result}
	}
}

func (a *App) renderHelp() string {
	title := lipgloss.NewStyle().Foreground(colorAccent).Bold(true).Render("devnews")
	dim := helpDimStyle

	help := title + dim.Render(" — Keyboard Shortcuts") + "\n\n" +
		dim.Render("Navigation") + "\n" +
		"  j/k, ↑/↓     Navigate article list\n" +
		"  tab           Switch focus between list and preview\n\n" +
		dim.Render("Actions") + "\n" +
		"  o, enter      Open article in browser\n" +
		"  r             Refresh feeds\n" +
		"  /             Search articles\n" +
		"  f             Toggle source filter mode\n\n" +
		dim.Render("Filter Mode") + "\n" +
		"  ←/→, h/l     Move between sources\n" +
		"  space/enter   Toggle source\n" +
		"  1-9           Toggle source by number\n" +
		"  esc, f        Exit filter mode\n\n" +
		dim.Render("General") + "\n" +
		"  h             Go to home screen\n" +
		"  ?             Toggle this help\n" +
		"  q, ctrl+c    Quit"

	card := helpCardStyle.Render(help)

	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, card)
}

// Run starts the TUI application.
func Run(opts RunOpts) error {
	app := NewApp(opts)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
