package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Cache struct {
	readDB  *sql.DB
	writeDB *sql.DB
}

func Open(dbPath string) (*Cache, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	writeDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening write db: %w", err)
	}
	writeDB.SetMaxOpenConns(1)

	readDB, err := sql.Open("sqlite", dbPath+"?mode=ro")
	if err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("opening read db: %w", err)
	}

	c := &Cache{readDB: readDB, writeDB: writeDB}
	if err := c.init(); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func (c *Cache) init() error {
	_, err := c.writeDB.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id          TEXT PRIMARY KEY,
			source      TEXT NOT NULL,
			title       TEXT NOT NULL,
			link        TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			published   DATETIME NOT NULL,
			fetched_at  DATETIME NOT NULL,
			summary     TEXT NOT NULL DEFAULT '',
			tags        TEXT NOT NULL DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS idx_articles_published ON articles(published DESC);
		CREATE INDEX IF NOT EXISTS idx_articles_source ON articles(source);

		CREATE TABLE IF NOT EXISTS meta (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("initializing schema: %w", err)
	}

	// Migrate: add summary/tags columns if they don't exist
	c.writeDB.Exec("ALTER TABLE articles ADD COLUMN summary TEXT NOT NULL DEFAULT ''")
	c.writeDB.Exec("ALTER TABLE articles ADD COLUMN tags TEXT NOT NULL DEFAULT ''")

	// Migrate: add V2 columns
	c.writeDB.Exec("ALTER TABLE articles ADD COLUMN signal_score REAL NOT NULL DEFAULT 0")
	c.writeDB.Exec("ALTER TABLE articles ADD COLUMN category TEXT NOT NULL DEFAULT ''")
	c.writeDB.Exec("ALTER TABLE articles ADD COLUMN why_it_matters TEXT NOT NULL DEFAULT ''")

	return nil
}

func (c *Cache) Close() error {
	var errs []error
	if c.readDB != nil {
		errs = append(errs, c.readDB.Close())
	}
	if c.writeDB != nil {
		errs = append(errs, c.writeDB.Close())
	}
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func (c *Cache) UpsertArticles(articles []Article) error {
	tx, err := c.writeDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO articles (id, source, title, link, description, published, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			fetched_at = excluded.fetched_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, a := range articles {
		_, err := stmt.Exec(a.ID, a.Source, a.Title, a.Link, a.Description, a.Published, a.FetchedAt)
		if err != nil {
			return fmt.Errorf("upserting article %s: %w", a.ID, err)
		}
	}

	return tx.Commit()
}

func (c *Cache) GetArticles(opts QueryOpts) ([]Article, error) {
	var (
		where []string
		args  []interface{}
	)

	if !opts.Since.IsZero() {
		where = append(where, "published >= ?")
		args = append(args, opts.Since)
	}

	if len(opts.Sources) > 0 {
		placeholders := make([]string, len(opts.Sources))
		for i, s := range opts.Sources {
			placeholders[i] = "?"
			args = append(args, s)
		}
		where = append(where, "source IN ("+strings.Join(placeholders, ",")+")") //nolint:gosec
	}

	if opts.Search != "" {
		where = append(where, "(title LIKE ? OR description LIKE ?)")
		term := "%" + opts.Search + "%"
		args = append(args, term, term)
	}

	if opts.Category != "" {
		where = append(where, "category = ?")
		args = append(args, opts.Category)
	}

	query := "SELECT id, source, title, link, description, published, fetched_at, summary, tags, category, why_it_matters FROM articles"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += " ORDER BY published DESC"

	limit := opts.Limit
	if limit <= 0 {
		limit = 500
	}
	query += fmt.Sprintf(" LIMIT %d", limit)

	rows, err := c.readDB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying articles: %w", err)
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Source, &a.Title, &a.Link, &a.Description, &a.Published, &a.FetchedAt, &a.Summary, &a.Tags, &a.Category, &a.WhyItMatters); err != nil {
			return nil, fmt.Errorf("scanning article: %w", err)
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

func (c *Cache) NeedsRefresh(interval time.Duration) bool {
	var value string
	err := c.readDB.QueryRow("SELECT value FROM meta WHERE key = 'last_refresh'").Scan(&value)
	if err != nil {
		return true
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return true
	}
	return time.Since(t) > interval
}

// Prune deletes articles older than the given retention duration and runs VACUUM.
// Returns the number of deleted rows.
func (c *Cache) Prune(retention time.Duration) (int64, error) {
	cutoff := time.Now().Add(-retention)
	result, err := c.writeDB.Exec("DELETE FROM articles WHERE published < ?", cutoff)
	if err != nil {
		return 0, fmt.Errorf("pruning articles: %w", err)
	}
	deleted, _ := result.RowsAffected()

	if deleted > 0 {
		if _, err := c.writeDB.Exec("VACUUM"); err != nil {
			return deleted, fmt.Errorf("vacuum after prune: %w", err)
		}
	}
	return deleted, nil
}

// Stats returns the total article count and database file size in bytes.
func (c *Cache) Stats(dbPath string) (count int64, sizeBytes int64, err error) {
	if err := c.readDB.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count); err != nil {
		return 0, 0, fmt.Errorf("counting articles: %w", err)
	}
	info, err := os.Stat(dbPath)
	if err != nil {
		return count, 0, nil // file might not exist yet
	}
	return count, info.Size(), nil
}

func (c *Cache) SetLastRefresh() error {
	_, err := c.writeDB.Exec(`
		INSERT INTO meta (key, value) VALUES ('last_refresh', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, time.Now().Format(time.RFC3339))
	return err
}

// UpdateArticleSummary saves a generated summary and tags for an article.
func (c *Cache) UpdateArticleSummary(id, summary, tags string) error {
	_, err := c.writeDB.Exec("UPDATE articles SET summary = ?, tags = ? WHERE id = ?", summary, tags, id)
	return err
}

func (c *Cache) setMeta(key, value string) error {
	_, err := c.writeDB.Exec(`
		INSERT INTO meta (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

func (c *Cache) getMeta(key string) (string, error) {
	var value string
	err := c.readDB.QueryRow("SELECT value FROM meta WHERE key = ?", key).Scan(&value)
	return value, err
}

// UpdateStreak updates the reading streak based on the current date.
// Returns the current streak count.
func (c *Cache) UpdateStreak() (int, error) {
	today := time.Now().Format("2006-01-02")

	lastDate, err := c.getMeta("last_active_date")
	if err != nil {
		// First launch ever
		c.setMeta("streak_days", "1")
		c.setMeta("last_active_date", today)
		return 1, nil
	}

	if lastDate == today {
		// Already counted today
		streak, _ := c.getMeta("streak_days")
		days := 1
		fmt.Sscanf(streak, "%d", &days)
		return days, nil
	}

	// Check if yesterday
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	streak := 1
	if lastDate == yesterday {
		old, _ := c.getMeta("streak_days")
		fmt.Sscanf(old, "%d", &streak)
		streak++
	}

	c.setMeta("streak_days", fmt.Sprintf("%d", streak))
	c.setMeta("last_active_date", today)
	return streak, nil
}

// GetLastOpened returns the last time devnews was opened.
func (c *Cache) GetLastOpened() (time.Time, error) {
	value, err := c.getMeta("last_opened")
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, value)
}

// SetLastOpened records the current time as the last opened time.
func (c *Cache) SetLastOpened() error {
	return c.setMeta("last_opened", time.Now().Format(time.RFC3339))
}

// GetArticlesSince returns articles published after the given time.
func (c *Cache) GetArticlesSince(since time.Time) ([]Article, error) {
	rows, err := c.readDB.Query(
		"SELECT id, source, title, link, description, published, fetched_at, summary, tags, category, why_it_matters FROM articles WHERE published >= ? ORDER BY published DESC",
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Source, &a.Title, &a.Link, &a.Description, &a.Published, &a.FetchedAt, &a.Summary, &a.Tags, &a.Category, &a.WhyItMatters); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, rows.Err()
}

// UpdateArticleCategory saves the category for an article.
func (c *Cache) UpdateArticleCategory(id, category string) error {
	_, err := c.writeDB.Exec("UPDATE articles SET category = ? WHERE id = ?", category, id)
	return err
}

// UpdateArticleWhyItMatters saves the "why it matters" text for an article.
func (c *Cache) UpdateArticleWhyItMatters(id, text string) error {
	_, err := c.writeDB.Exec("UPDATE articles SET why_it_matters = ? WHERE id = ?", text, id)
	return err
}

// ShouldCheckUpdate returns true if the last update check was more than 24 hours ago or never happened.
func (c *Cache) ShouldCheckUpdate() bool {
	value, err := c.getMeta("last_update_check")
	if err != nil {
		return true
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return true
	}
	return time.Since(t) > 24*time.Hour
}

// SetLastUpdateCheck records the current time as the last update check.
func (c *Cache) SetLastUpdateCheck() error {
	return c.setMeta("last_update_check", time.Now().Format(time.RFC3339))
}

