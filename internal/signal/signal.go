package signal

import (
	"math"
	"strings"
	"time"
	"unicode"
)

// SourceWeights maps source names to their weight (0.0–1.0).
type SourceWeights map[string]float64

// Input holds the data needed to score an article.
type Input struct {
	Title       string
	Description string
	Source      string
	Published   time.Time
}

// Breakdown shows how each component contributed to the final score.
type Breakdown struct {
	Recency        float64
	SourceWeight   float64
	Depth          float64
	KeywordDensity float64
	Final          float64
}

const (
	weightRecency  = 0.30
	weightSource   = 0.25
	weightDepth    = 0.25
	weightKeywords = 0.20
)

// Score computes a signal score (0.0–10.0) for an article.
func Score(input Input, weights SourceWeights) float64 {
	return ScoreWithBreakdown(input, weights).Final
}

// ScoreWithBreakdown computes a signal score with component details.
func ScoreWithBreakdown(input Input, weights SourceWeights) Breakdown {
	b := Breakdown{
		Recency:        recencyScore(input.Published),
		SourceWeight:   sourceScore(input.Source, weights),
		Depth:          depthScore(input.Description),
		KeywordDensity: keywordScore(input.Title, input.Description),
	}
	raw := b.Recency*weightRecency +
		b.SourceWeight*weightSource +
		b.Depth*weightDepth +
		b.KeywordDensity*weightKeywords
	b.Final = math.Round(raw*100) / 10 // scale to 0.0–10.0
	return b
}

// recencyScore returns exponential decay: 1.0 at publish, ~0.5 at 24h, ~0.1 at 72h.
func recencyScore(published time.Time) float64 {
	if published.IsZero() {
		return 0.0
	}
	hours := time.Since(published).Hours()
	if hours < 0 {
		hours = 0
	}
	// decay constant: ln(0.5)/24 ≈ -0.02888
	return math.Exp(-0.02888 * hours)
}

// sourceScore looks up the source weight, defaulting to 0.5.
func sourceScore(source string, weights SourceWeights) float64 {
	if weights == nil {
		return 0.5
	}
	if w, ok := weights[source]; ok {
		return w
	}
	return 0.5
}

// depthScore scores based on description word count.
func depthScore(description string) float64 {
	words := len(strings.Fields(description))
	switch {
	case words >= 150:
		return 1.0
	case words >= 50:
		return 0.6
	default:
		return 0.2
	}
}

// engineeringKeywords are high-signal terms for engineering content.
var engineeringKeywords = map[string]bool{
	"scale": true, "scaling": true, "performance": true, "latency": true,
	"throughput": true, "reliability": true, "resilience": true,
	"architecture": true, "microservice": true, "distributed": true,
	"consensus": true, "replication": true, "sharding": true,
	"kubernetes": true, "container": true, "docker": true,
	"rust": true, "golang": true, "typescript": true,
	"database": true, "cache": true, "index": true,
	"encryption": true, "authentication": true, "security": true,
	"inference": true, "training": true, "model": true,
	"pipeline": true, "deployment": true, "observability": true,
	"migration": true, "optimization": true, "concurrency": true,
	"fault": true, "tolerant": true, "idempotent": true,
	"streaming": true, "realtime": true, "async": true,
}

// keywordScore returns the density of engineering keywords (0.0–1.0).
func keywordScore(title, description string) float64 {
	text := strings.ToLower(title + " " + description)
	var words []string
	for _, w := range strings.Fields(text) {
		w = strings.TrimFunc(w, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		if w != "" {
			words = append(words, w)
		}
	}
	if len(words) == 0 {
		return 0.0
	}

	hits := 0
	for _, w := range words {
		if engineeringKeywords[w] {
			hits++
		}
	}
	density := float64(hits) / float64(len(words))
	// Normalize: 10%+ keyword density = 1.0
	score := density * 10
	if score > 1.0 {
		score = 1.0
	}
	return score
}
