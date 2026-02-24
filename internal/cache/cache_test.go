package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testDB(t *testing.T) *Cache {
	t.Helper()
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func sampleArticles() []Article {
	now := time.Now()
	return []Article{
		{ID: "aaa", Source: "Cloudflare", Title: "Post A", Link: "https://a.com", Description: "Desc A", Published: now.Add(-1 * time.Hour), FetchedAt: now},
		{ID: "bbb", Source: "GitHub", Title: "Post B", Link: "https://b.com", Description: "Desc B", Published: now.Add(-2 * time.Hour), FetchedAt: now},
		{ID: "ccc", Source: "Cloudflare", Title: "Post C", Link: "https://c.com", Description: "Desc C about search", Published: now.Add(-48 * time.Hour), FetchedAt: now},
	}
}

func TestUpsertAndGet(t *testing.T) {
	db := testDB(t)
	articles := sampleArticles()

	if err := db.UpsertArticles(articles); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetArticles(QueryOpts{})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 articles, got %d", len(got))
	}
	// Should be ordered by published DESC
	if got[0].ID != "aaa" {
		t.Errorf("expected newest first, got %s", got[0].ID)
	}
}

func TestUpsertUpdatesExisting(t *testing.T) {
	db := testDB(t)
	articles := sampleArticles()

	if err := db.UpsertArticles(articles); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	// Update title
	articles[0].Title = "Updated Post A"
	if err := db.UpsertArticles(articles[:1]); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	got, err := db.GetArticles(QueryOpts{})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 articles after upsert, got %d", len(got))
	}
	if got[0].Title != "Updated Post A" {
		t.Errorf("expected updated title, got %q", got[0].Title)
	}
}

func TestQuerySince(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetArticles(QueryOpts{Since: time.Now().Add(-3 * time.Hour)})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 articles within 3h, got %d", len(got))
	}
}

func TestQuerySources(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetArticles(QueryOpts{Sources: []string{"Cloudflare"}})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 Cloudflare articles, got %d", len(got))
	}
	for _, a := range got {
		if a.Source != "Cloudflare" {
			t.Errorf("expected source Cloudflare, got %s", a.Source)
		}
	}
}

func TestQuerySearch(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetArticles(QueryOpts{Search: "search"})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 article matching 'search', got %d", len(got))
	}
	if len(got) > 0 && got[0].ID != "ccc" {
		t.Errorf("expected article ccc, got %s", got[0].ID)
	}
}

func TestQueryCombinedFilters(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Cloudflare + within 3h = only Post A
	got, err := db.GetArticles(QueryOpts{
		Sources: []string{"Cloudflare"},
		Since:   time.Now().Add(-3 * time.Hour),
	})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 article, got %d", len(got))
	}
}

func TestQueryLimit(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetArticles(QueryOpts{Limit: 1})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 article with limit, got %d", len(got))
	}
}

func TestNeedsRefresh(t *testing.T) {
	db := testDB(t)

	// No last_refresh set — should need refresh
	if !db.NeedsRefresh(1 * time.Hour) {
		t.Error("expected NeedsRefresh=true when no last_refresh set")
	}

	// Set last refresh
	if err := db.SetLastRefresh(); err != nil {
		t.Fatalf("SetLastRefresh: %v", err)
	}

	// Just refreshed — should not need refresh
	if db.NeedsRefresh(1 * time.Hour) {
		t.Error("expected NeedsRefresh=false right after SetLastRefresh")
	}

	// With zero interval — should always need refresh
	if !db.NeedsRefresh(0) {
		t.Error("expected NeedsRefresh=true with zero interval")
	}
}

func TestEmptyDB(t *testing.T) {
	db := testDB(t)

	got, err := db.GetArticles(QueryOpts{})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 articles in empty db, got %d", len(got))
	}
}

func TestPruneDeletesOldArticles(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Post C is 48h old. Prune anything older than 24h.
	deleted, err := db.Prune(24 * time.Hour)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 pruned, got %d", deleted)
	}

	got, err := db.GetArticles(QueryOpts{})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 remaining articles, got %d", len(got))
	}
}

func TestPruneNothingToDelete(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	deleted, err := db.Prune(365 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 pruned, got %d", deleted)
	}
}

func TestStats(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	count, size, err := db.Stats(dbPath)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
	if size == 0 {
		t.Error("expected non-zero db size")
	}
}

func TestStreakFirstLaunch(t *testing.T) {
	db := testDB(t)
	streak, err := db.UpdateStreak()
	if err != nil {
		t.Fatalf("UpdateStreak: %v", err)
	}
	if streak != 1 {
		t.Errorf("expected streak 1 on first launch, got %d", streak)
	}
}

func TestStreakSameDay(t *testing.T) {
	db := testDB(t)
	db.UpdateStreak()
	streak, _ := db.UpdateStreak()
	if streak != 1 {
		t.Errorf("expected streak 1 on same day, got %d", streak)
	}
}

func TestStreakNextDay(t *testing.T) {
	db := testDB(t)
	// Simulate yesterday
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	db.setMeta("last_active_date", yesterday)
	db.setMeta("streak_days", "5")

	streak, _ := db.UpdateStreak()
	if streak != 6 {
		t.Errorf("expected streak 6, got %d", streak)
	}
}

func TestStreakReset(t *testing.T) {
	db := testDB(t)
	// Simulate 3 days ago
	old := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	db.setMeta("last_active_date", old)
	db.setMeta("streak_days", "10")

	streak, _ := db.UpdateStreak()
	if streak != 1 {
		t.Errorf("expected streak reset to 1, got %d", streak)
	}
}

func TestLastOpened(t *testing.T) {
	db := testDB(t)

	// No last opened initially
	_, err := db.GetLastOpened()
	if err == nil {
		t.Error("expected error when no last_opened set")
	}

	// Set and retrieve
	db.SetLastOpened()
	got, err := db.GetLastOpened()
	if err != nil {
		t.Fatalf("GetLastOpened: %v", err)
	}
	if time.Since(got) > 2*time.Second {
		t.Errorf("last opened too old: %v", got)
	}
}

func TestUpdateArticleSummary(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if err := db.UpdateArticleSummary("aaa", "Test summary", "rust, dns"); err != nil {
		t.Fatalf("UpdateArticleSummary: %v", err)
	}

	articles, _ := db.GetArticles(QueryOpts{})
	for _, a := range articles {
		if a.ID == "aaa" {
			if a.Summary != "Test summary" {
				t.Errorf("expected summary 'Test summary', got %q", a.Summary)
			}
			if a.Tags != "rust, dns" {
				t.Errorf("expected tags 'rust, dns', got %q", a.Tags)
			}
			return
		}
	}
	t.Error("article aaa not found")
}

func TestGetArticlesSince(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := db.GetArticlesSince(time.Now().Add(-3 * time.Hour))
	if err != nil {
		t.Fatalf("GetArticlesSince: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 articles within 3h, got %d", len(got))
	}
}

func TestUpdateArticleSignal(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if err := db.UpdateArticleSignal("aaa", 8.7, "Infrastructure"); err != nil {
		t.Fatalf("UpdateArticleSignal: %v", err)
	}

	articles, _ := db.GetArticles(QueryOpts{})
	for _, a := range articles {
		if a.ID == "aaa" {
			if a.SignalScore != 8.7 {
				t.Errorf("expected signal score 8.7, got %.1f", a.SignalScore)
			}
			if a.Category != "Infrastructure" {
				t.Errorf("expected category Infrastructure, got %q", a.Category)
			}
			return
		}
	}
	t.Error("article aaa not found")
}

func TestUpdateArticleWhyItMatters(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	if err := db.UpdateArticleWhyItMatters("bbb", "This matters because..."); err != nil {
		t.Fatalf("UpdateArticleWhyItMatters: %v", err)
	}

	articles, _ := db.GetArticles(QueryOpts{})
	for _, a := range articles {
		if a.ID == "bbb" {
			if a.WhyItMatters != "This matters because..." {
				t.Errorf("expected why_it_matters text, got %q", a.WhyItMatters)
			}
			return
		}
	}
	t.Error("article bbb not found")
}

func TestSignalOrdering(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Set different signal scores
	db.UpdateArticleSignal("aaa", 5.0, "")
	db.UpdateArticleSignal("bbb", 9.0, "")
	db.UpdateArticleSignal("ccc", 7.0, "")

	got, err := db.GetArticles(QueryOpts{OrderBy: "signal"})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 articles, got %d", len(got))
	}
	if got[0].ID != "bbb" {
		t.Errorf("expected highest signal first (bbb), got %s", got[0].ID)
	}
}

func TestCategoryFilter(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	db.UpdateArticleSignal("aaa", 8.0, "Infrastructure")
	db.UpdateArticleSignal("bbb", 7.0, "AI/ML")
	db.UpdateArticleSignal("ccc", 6.0, "Infrastructure")

	got, err := db.GetArticles(QueryOpts{Category: "Infrastructure"})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 Infrastructure articles, got %d", len(got))
	}
}

func TestGetTopArticles(t *testing.T) {
	db := testDB(t)
	if err := db.UpsertArticles(sampleArticles()); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	db.UpdateArticleSignal("aaa", 5.0, "")
	db.UpdateArticleSignal("bbb", 9.0, "")
	db.UpdateArticleSignal("ccc", 7.0, "")

	got, err := db.GetTopArticles(time.Now().Add(-72*time.Hour), 2, "")
	if err != nil {
		t.Fatalf("GetTopArticles: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 top articles, got %d", len(got))
	}
	if len(got) > 0 && got[0].ID != "bbb" {
		t.Errorf("expected bbb first, got %s", got[0].ID)
	}
}

func TestOpenCreatesDir(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "sub", "deep", "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("opening db in nested dir: %v", err)
	}
	db.Close()

	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("expected directory to be created")
	}
}
