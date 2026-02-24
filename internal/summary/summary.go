package summary

import "github.com/matheuskafuri/devnews/internal/cache"

// Summarizer generates a short summary for an article.
// Post-MVP: implement with Claude/OpenAI API.
type Summarizer interface {
	Summarize(article cache.Article) (string, error)
}

// Tagger assigns topic tags to an article.
// Post-MVP: implement with LLM or keyword extraction.
type Tagger interface {
	Tag(article cache.Article) ([]string, error)
}
