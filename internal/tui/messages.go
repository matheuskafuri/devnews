package tui

import (
	"github.com/matheuskafuri/devnews/internal/ai"
	"github.com/matheuskafuri/devnews/internal/cache"
)

type feedsLoadedMsg struct {
	articles []cache.Article
}

type feedErrMsg struct {
	err error
}

type refreshDoneMsg struct {
	count int
	errs  []error
}

type summaryLoadedMsg struct {
	articleID string
	result    ai.Result
}

type whyItMattersMsg struct {
	cardIndex int
	articleID string
	text      string
}

type themesMsg struct {
	themes []string
}

type updateAvailableMsg struct {
	version string
}

