package cache

import "time"

type Article struct {
	ID            string
	Source        string
	Title         string
	Link          string
	Description   string
	Published     time.Time
	FetchedAt     time.Time
	Summary       string
	Tags          string
	SignalScore   float64
	Category      string
	WhyItMatters  string
}

type QueryOpts struct {
	Since    time.Time
	Sources  []string
	Search   string
	Limit    int
	Category string
	OrderBy  string // "published" (default) or "signal"
}
