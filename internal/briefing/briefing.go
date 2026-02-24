package briefing

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/matheuskafuri/devnews/internal/cache"
	"github.com/matheuskafuri/devnews/internal/classify"
	"github.com/matheuskafuri/devnews/internal/signal"
)

// Briefing holds briefing data for both V2 (card-based) and V1 (legacy header) modes.
type Briefing struct {
	// V2 fields
	DateLabel string
	Freshness string
	Scanned   int
	Selected  int
	Themes    []string
	Cards     []Card
	Focus     string

	// V1 fields (used by legacy browse-mode header)
	Greeting      string
	NewCount      int
	ActiveSources string
	Trending      string
}

// Card represents a single briefing card.
type Card struct {
	Article     cache.Article
	Index       int
	ReadingTime int
}

// GenerateOpts holds options for the Generate function.
type GenerateOpts struct {
	DB            *cache.Cache
	Since         time.Time
	BriefSize     int
	FocusCategory string
	SourceWeights map[string]float64
}

// Generate creates a V2 briefing by scoring, classifying, and selecting top articles.
// AI enrichment (WhyItMatters, Themes) is NOT done here — the caller should
// handle those asynchronously so the opening screen renders immediately.
func Generate(opts GenerateOpts) (*Briefing, error) {
	if opts.BriefSize <= 0 {
		opts.BriefSize = 5
	}

	// Fetch all articles in the window
	articles, err := opts.DB.GetArticlesSince(opts.Since)
	if err != nil {
		return nil, fmt.Errorf("fetching articles: %w", err)
	}

	b := &Briefing{
		DateLabel: time.Now().Format("Jan 2"),
		Freshness: "Fresh",
		Scanned:   len(articles),
		Focus:     opts.FocusCategory,
	}

	if len(articles) == 0 {
		return b, nil
	}

	// Score and classify each article
	for i := range articles {
		input := signal.Input{
			Title:       articles[i].Title,
			Description: articles[i].Description,
			Source:      articles[i].Source,
			Published:   articles[i].Published,
		}
		articles[i].SignalScore = signal.Score(input, opts.SourceWeights)
		articles[i].Category = string(classify.Classify(articles[i].Title, articles[i].Description))

		// Persist to cache (non-blocking, best-effort)
		opts.DB.UpdateArticleSignal(articles[i].ID, articles[i].SignalScore, articles[i].Category)
	}

	// Sort by signal score descending
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].SignalScore > articles[j].SignalScore
	})

	// Filter by focus category if set
	if opts.FocusCategory != "" {
		var filtered []cache.Article
		for _, a := range articles {
			if a.Category == opts.FocusCategory {
				filtered = append(filtered, a)
			}
		}
		articles = filtered
	}

	// Take top N
	if len(articles) > opts.BriefSize {
		articles = articles[:opts.BriefSize]
	}

	b.Selected = len(articles)

	// Build cards
	for i, a := range articles {
		b.Cards = append(b.Cards, Card{
			Article:     a,
			Index:       i + 1,
			ReadingTime: estimateReadTime(a.Description),
		})
	}

	// Set description excerpt as initial "why it matters" (AI can override later)
	for i := range b.Cards {
		if b.Cards[i].Article.WhyItMatters == "" {
			b.Cards[i].Article.WhyItMatters = DescriptionExcerpt(b.Cards[i].Article.Description)
		}
	}

	// TF-IDF fallback for themes (AI themes loaded async by the TUI)
	if len(b.Themes) == 0 {
		allArticles, _ := opts.DB.GetArticles(cache.QueryOpts{})
		trendingText := trending(articles, allArticles)
		if trendingText != "" {
			b.Themes = strings.Split(trendingText, ", ")
		}
	}

	return b, nil
}

// GenerateLegacy creates a V1-style briefing (for backward compatibility).
func GenerateLegacy(newArticles []cache.Article, allArticles []cache.Article) Briefing {
	b := Briefing{
		Greeting: greeting(time.Now()),
		NewCount: len(newArticles),
	}

	if len(newArticles) > 0 {
		b.ActiveSources = activeSources(newArticles)
		b.Trending = trending(newArticles, allArticles)
	}

	return b
}

// DescriptionExcerpt returns the first sentence of a description as a fallback.
func DescriptionExcerpt(desc string) string {
	if desc == "" {
		return ""
	}
	// Take first sentence
	for i, c := range desc {
		if c == '.' && i > 20 {
			return desc[:i+1]
		}
	}
	runes := []rune(desc)
	if len(runes) > 150 {
		return string(runes[:150]) + "..."
	}
	return desc
}

func estimateReadTime(desc string) int {
	words := len(strings.Fields(desc))
	// Multiply by 3 for full article estimate, divide by 200 WPM
	minutes := (words * 3) / 200
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

func greeting(now time.Time) string {
	hour := now.Hour()
	switch {
	case hour < 12:
		return "Good morning"
	case hour < 17:
		return "Good afternoon"
	default:
		return "Good evening"
	}
}

func activeSources(articles []cache.Article) string {
	counts := map[string]int{}
	for _, a := range articles {
		counts[a.Source]++
	}

	type sc struct {
		name  string
		count int
	}
	var sorted []sc
	for name, count := range counts {
		sorted = append(sorted, sc{name, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	limit := 3
	if len(sorted) < limit {
		limit = len(sorted)
	}

	parts := make([]string, limit)
	for i := 0; i < limit; i++ {
		parts[i] = fmt.Sprintf("%s (%d)", sorted[i].name, sorted[i].count)
	}
	return strings.Join(parts, ", ")
}

// trending extracts top keywords from new article titles using TF-IDF.
func trending(newArticles []cache.Article, allArticles []cache.Article) string {
	df := map[string]int{}
	for _, a := range allArticles {
		seen := map[string]bool{}
		for _, w := range tokenize(a.Title) {
			if !seen[w] {
				df[w]++
				seen[w] = true
			}
		}
	}

	tf := map[string]int{}
	for _, a := range newArticles {
		for _, w := range tokenize(a.Title) {
			tf[w]++
		}
	}

	totalDocs := len(allArticles)
	if totalDocs == 0 {
		totalDocs = 1
	}

	type scored struct {
		term  string
		score float64
	}
	var terms []scored
	for term, freq := range tf {
		if freq < 2 {
			continue
		}
		docFreq := df[term]
		if docFreq == 0 {
			docFreq = 1
		}
		idf := math.Log(float64(totalDocs) / float64(docFreq))
		terms = append(terms, scored{term, float64(freq) * idf})
	}

	sort.Slice(terms, func(i, j int) bool {
		return terms[i].score > terms[j].score
	})

	limit := 3
	if len(terms) < limit {
		limit = len(terms)
	}

	parts := make([]string, limit)
	for i := 0; i < limit; i++ {
		parts[i] = terms[i].term
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ", ")
}

var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true, "is": true, "it": true, "its": true,
	"this": true, "that": true, "are": true, "was": true, "were": true, "be": true,
	"been": true, "being": true, "have": true, "has": true, "had": true, "do": true,
	"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
	"may": true, "might": true, "can": true, "not": true, "no": true, "nor": true,
	"how": true, "what": true, "when": true, "where": true, "who": true, "which": true,
	"why": true, "all": true, "each": true, "every": true, "both": true, "few": true,
	"more": true, "most": true, "other": true, "some": true, "such": true, "than": true,
	"too": true, "very": true, "just": true, "about": true, "into": true, "over": true,
	"after": true, "before": true, "between": true, "under": true, "above": true,
	"out": true, "up": true, "down": true, "off": true, "our": true, "your": true,
	"we": true, "you": true, "they": true, "them": true, "their": true, "new": true,
	"use": true, "using": true, "used": true,
}

func tokenize(s string) []string {
	var tokens []string
	for _, word := range strings.Fields(strings.ToLower(s)) {
		word = strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		if len(word) < 4 {
			continue
		}
		if stopWords[word] {
			continue
		}
		tokens = append(tokens, word)
	}
	return tokens
}
