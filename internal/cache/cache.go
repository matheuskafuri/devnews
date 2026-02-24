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
			fetched_at  DATETIME NOT NULL
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

	query := "SELECT id, source, title, link, description, published, fetched_at FROM articles"
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
		if err := rows.Scan(&a.ID, &a.Source, &a.Title, &a.Link, &a.Description, &a.Published, &a.FetchedAt); err != nil {
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

func (c *Cache) SetLastRefresh() error {
	_, err := c.writeDB.Exec(`
		INSERT INTO meta (key, value) VALUES ('last_refresh', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, time.Now().Format(time.RFC3339))
	return err
}
