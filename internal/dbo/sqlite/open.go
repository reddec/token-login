package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/reddec/token-login/internal/dbo"
	migrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite"
)

// Open opens a SQLite database, runs migrations, and returns a Store.
//
// The URL must use the sqlite, sqlite3, or file scheme. In-memory databases
// (":memory:") are supported; foreign keys are enabled after migration.
func Open(ctx context.Context, rawURL string, hook func(db *sql.DB)) (dbo.Store, error) {
	inMemory := strings.Contains(rawURL, ":memory:")

	connURL, err := prepareURL(rawURL, "foreign_keys(0)")
	if err != nil {
		return nil, fmt.Errorf("prepare sqlite URL: %w", err)
	}

	db, err := sql.Open("sqlite", connURL)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	closeDB := func() { _ = db.Close() }

	if hook != nil {
		hook(db)
	}

	if err := runMigrations(ctx, db); err != nil {
		closeDB()
		return nil, fmt.Errorf("run sqlite migrations: %w", err)
	}

	if inMemory {
		if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
			closeDB()
			return nil, fmt.Errorf("enable foreign keys: %w", err)
		}
		return NewStore(db), nil
	}

	closeDB()

	// Re-open with foreign keys enabled for file-based databases.
	connURL, err = prepareURL(rawURL, "foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("prepare sqlite runtime URL: %w", err)
	}
	runtimeDB, err := sql.Open("sqlite", connURL)
	if err != nil {
		return nil, fmt.Errorf("re-open sqlite: %w", err)
	}
	if hook != nil {
		hook(runtimeDB)
	}

	return NewStore(runtimeDB), nil
}

// prepareURL normalizes a SQLite DSN into the connection string that
// modernc.org/sqlite expects, applying the given pragma.
func prepareURL(rawURL, fkPragma string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		if !strings.Contains(rawURL, ":memory:") {
			return "", fmt.Errorf("parse DSN: %w", err)
		}
		u = new(url.URL)
		if before, _, ok := strings.Cut(rawURL, "://"); ok {
			u.Scheme = before
		}
		u.Host = ":memory:"
		if _, after, ok := strings.Cut(rawURL, "?"); ok {
			u.RawQuery = after
		}
	}
	u.Scheme = "file"
	q := u.Query()
	q.Set("_pragma", fkPragma)
	u.RawQuery = q.Encode()
	return strings.ReplaceAll(u.String(), "file://", "file:"), nil
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	source := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: MigrationsFS,
		Root:       "migrations",
	}

	// SQLite lacks ALTER TABLE ADD COLUMN IF NOT EXISTS, so if the host column
	// already exists on the token table (e.g. from an Ent v1.2.0 migration),
	// seed the tracking record to prevent a "duplicate column" error.
	seedHostMigrationIfNeeded(ctx, db)

	n, err := migrate.Exec(db, "sqlite3", source, migrate.Up)
	if err != nil {
		return fmt.Errorf("exec migrations: %w", err)
	}
	if n > 0 {
		slog.Info("applied sqlite migrations", "count", n)
	}
	return nil
}

// seedHostMigrationIfNeeded detects whether the host column already exists on
// the token table and, if so, seeds the gorp_migrations tracking table so that
// migration 002_add_host.sql is skipped.
func seedHostMigrationIfNeeded(ctx context.Context, db *sql.DB) {
	var hasHost bool
	_ = db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM pragma_table_info('token') WHERE name='host')",
	).Scan(&hasHost)
	if !hasHost {
		return
	}

	_, _ = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS gorp_migrations (id TEXT NOT NULL PRIMARY KEY, applied_at DATETIME)`)

	var tracked bool
	_ = db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM gorp_migrations WHERE id='002_add_host.sql')",
	).Scan(&tracked)
	if tracked {
		return
	}

	_, _ = db.ExecContext(ctx,
		"INSERT INTO gorp_migrations (id, applied_at) VALUES ('002_add_host.sql', ?)",
		time.Now())
	slog.Info("sqlite: seeded migration 002_add_host.sql (host column already present)")
}
