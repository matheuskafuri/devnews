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
