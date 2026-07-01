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
// (":memory:") are supported. Foreign keys are enabled via the DSN so every
// pooled connection inherits the setting.
func Open(ctx context.Context, rawURL string, hook func(db *sql.DB)) (dbo.Store, error) {
	connURL, err := prepareURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("prepare sqlite URL: %w", err)
	}

	db, err := sql.Open("sqlite", connURL)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if hook != nil {
		hook(db)
	}

	if err := runMigrations(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("run sqlite migrations: %w", err)
	}

	return NewStore(db), nil
}

// prepareURL normalizes a SQLite DSN into the connection string that
// modernc.org/sqlite expects, with foreign keys and shared cache enabled.
func prepareURL(rawURL string) (string, error) {
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
	q.Set("cache", "shared")
	q.Set("_pragma", "foreign_keys(1)")
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

	n, err := migrate.ExecContext(ctx, db, "sqlite3", source, migrate.Up)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
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
	if err := db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM pragma_table_info('token') WHERE name='host')",
	).Scan(&hasHost); err != nil {
		slog.Warn("sqlite: failed to check host column presence for migration seed", "error", err)
		return
	}
	if !hasHost {
		return
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS gorp_migrations (id TEXT NOT NULL PRIMARY KEY, applied_at DATETIME)`); err != nil {
		slog.Warn("sqlite: failed to create gorp_migrations table for migration seed", "error", err)
		return
	}

	var tracked bool
	if err := db.QueryRowContext(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM gorp_migrations WHERE id='002_add_host.sql')",
	).Scan(&tracked); err != nil {
		slog.Warn("sqlite: failed to check migration tracking for seed", "error", err)
		return
	}
	if tracked {
		return
	}

	if _, err := db.ExecContext(ctx,
		"INSERT INTO gorp_migrations (id, applied_at) VALUES ('002_add_host.sql', ?)",
		time.Now()); err != nil {
		slog.Warn("sqlite: failed to seed migration 002_add_host.sql", "error", err)
		return
	}
	slog.Info("sqlite: seeded migration 002_add_host.sql (host column already present)")
}
