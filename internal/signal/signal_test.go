package signal

import (
	"math"
	"testing"
	"time"
)

func TestScoreRecentArticle(t *testing.T) {
	input := Input{
		Title:       "Scaling Kubernetes Clusters for High Throughput",
		Description: "We describe our approach to scaling container orchestration across multiple regions with performance optimization and reliability improvements. " + longDescription(),
		Source:      "Cloudflare",
		Published:   time.Now(),
	}
	weights := SourceWeights{"Cloudflare": 0.9}

	score := Score(input, weights)
	if score < 5.0 {
		t.Errorf("expected high score for recent, good-source article, got %.1f", score)
	}
	if score > 10.0 {
		t.Errorf("score should not exceed 10.0, got %.1f", score)
	}
}

func TestScoreOldArticle(t *testing.T) {
	input := Input{
		Title:       "Scaling Kubernetes Clusters",
		Description: longDescription(),
		Source:      "Cloudflare",
		Published:   time.Now().Add(-72 * time.Hour),
	}
	weights := SourceWeights{"Cloudflare": 0.9}

	score := Score(input, weights)
	if score > 7.0 {
		t.Errorf("expected lower score for 72h old article, got %.1f", score)
	}
}

func TestRecencyDecay(t *testing.T) {
	now := recencyScore(time.Now())
	day := recencyScore(time.Now().Add(-24 * time.Hour))
	threeDay := recencyScore(time.Now().Add(-72 * time.Hour))

	if now < 0.95 {
		t.Errorf("recency now should be ~1.0, got %.2f", now)
	}
	if math.Abs(day-0.5) > 0.1 {
		t.Errorf("recency at 24h should be ~0.5, got %.2f", day)
	}
	if threeDay > 0.2 {
		t.Errorf("recency at 72h should be <0.2, got %.2f", threeDay)
	}
}

func TestDefaultSourceWeight(t *testing.T) {
	score := sourceScore("Unknown", nil)
	if score != 0.5 {
		t.Errorf("expected default 0.5, got %.2f", score)
	}

	score = sourceScore("Unknown", SourceWeights{"Other": 0.9})
	if score != 0.5 {
		t.Errorf("expected default 0.5 for missing source, got %.2f", score)
	}
}

func TestDepthScoreBands(t *testing.T) {
	short := depthScore("hello world")
	if short != 0.2 {
		t.Errorf("expected 0.2 for short, got %.1f", short)
	}

	medium := depthScore(nWords(75))
	if medium != 0.6 {
		t.Errorf("expected 0.6 for medium, got %.1f", medium)
	}

	long := depthScore(nWords(200))
	if long != 1.0 {
		t.Errorf("expected 1.0 for long, got %.1f", long)
	}
}

func TestBreakdownComponents(t *testing.T) {
	input := Input{
		Title:       "Distributed Database Replication",
		Description: longDescription(),
		Source:      "Stripe",
		Published:   time.Now(),
	}
	weights := SourceWeights{"Stripe": 0.8}

	b := ScoreWithBreakdown(input, weights)
	if b.Recency < 0.9 {
		t.Errorf("recency should be high for fresh article, got %.2f", b.Recency)
	}
	if b.SourceWeight != 0.8 {
		t.Errorf("source weight should be 0.8, got %.2f", b.SourceWeight)
	}
	if b.Final < 0 || b.Final > 10 {
		t.Errorf("final score out of range: %.1f", b.Final)
	}
}

func TestScoreZeroInput(t *testing.T) {
	score := Score(Input{}, nil)
	if score < 0 || score > 10 {
		t.Errorf("score out of range for zero input: %.1f", score)
	}
}

func longDescription() string {
	s := ""
	for i := 0; i < 160; i++ {
		s += "word "
	}
	return s
}

func nWords(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "word "
	}
	return s
}
