package feed

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/matheuskafuri/devnews/internal/cache"
	"github.com/matheuskafuri/devnews/internal/config"
	"github.com/mmcdole/gofeed"
)

type Fetcher interface {
	Fetch(ctx context.Context, source config.Source) ([]cache.Article, error)
}

type RSSFetcher struct {
	parser *gofeed.Parser
}

func NewRSSFetcher() *RSSFetcher {
	return &RSSFetcher{parser: gofeed.NewParser()}
}

func (f *RSSFetcher) Fetch(ctx context.Context, source config.Source) ([]cache.Article, error) {
	feed, err := f.parser.ParseURLWithContext(source.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", source.Name, err)
	}

	now := time.Now()
	maxAge := now.Add(-7 * 24 * time.Hour)
	articles := make([]cache.Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		pub := now
		if item.PublishedParsed != nil {
			pub = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			pub = *item.UpdatedParsed
		}

		// Skip articles older than 7 days
		if pub.Before(maxAge) {
			continue
		}

		desc := item.Description
		if desc == "" {
			desc = item.Content
		}
		desc = truncate(stripHTML(desc), 300)

		articles = append(articles, cache.Article{
			ID:          articleID(item.Link),
			Source:      source.Name,
			Title:       item.Title,
			Link:        item.Link,
			Description: desc,
			Published:   pub,
			FetchedAt:   now,
		})
	}
	return articles, nil
}

func articleID(link string) string {
	h := sha256.Sum256([]byte(link))
	return fmt.Sprintf("%x", h[:16])
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 3 {
		return string(runes[:n])
	}
	return string(runes[:n-3]) + "..."
}

func stripHTML(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

type FetchResult struct {
	Articles []cache.Article
	Errors   []error
}

func FetchAll(ctx context.Context, sources []config.Source) FetchResult {
	var (
		mu     sync.Mutex
		result FetchResult
		wg     sync.WaitGroup
	)

	fetcher := NewRSSFetcher()

	for _, src := range sources {
		wg.Add(1)
		go func(s config.Source) {
			defer wg.Done()
			articles, err := fetcher.Fetch(ctx, s)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				result.Errors = append(result.Errors, err)
				return
			}
			result.Articles = append(result.Articles, articles...)
		}(src)
	}

	wg.Wait()
	return result
}
