package tui

import (
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
