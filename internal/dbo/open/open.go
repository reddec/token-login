// Package open provides the database initialization factory.
// It handles opening connections, running migrations, and wiring adapters.
package open

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver for database/sql
	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/internal/dbo/postgres"
	"github.com/reddec/token-login/internal/dbo/sqlite"
	sqlmigrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite" // SQLite driver for database/sql
)

var errUnsupportedScheme = errors.New("unsupported database scheme")

// Open opens a database connection based on the URL scheme, runs pending
// migrations, and returns a Store implementation.
//
// Supported schemes:
//   - sqlite, sqlite3, file  → SQLite via modernc.org/sqlite
//   - postgres               → PostgreSQL via pgx/v5
func Open(ctx context.Context, rawURL string, hook func(db *sql.DB)) (dbo.Store, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		// url.Parse rejects non-standard authorities like ":memory:" (treated
		// as invalid port). For in-memory SQLite DSNs, construct the URL
		// manually so openSQLite can transform it into the file::memory:…
		// connection string that modernc.org/sqlite expects.
		if !strings.Contains(rawURL, ":memory:") {
			return nil, fmt.Errorf("parse DSN: %w", err)
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
	switch u.Scheme {
	case "sqlite", "sqlite3", "file":
		return openSQLite(ctx, u, rawURL, hook)
	case "postgres":
		return openPostgres(ctx, u, hook)
	default:
		return nil, fmt.Errorf("unsupported database scheme %s: %w", u.Scheme, errUnsupportedScheme)
	}
}

func openSQLite(ctx context.Context, u *url.URL, rawURL string, hook func(db *sql.DB)) (dbo.Store, error) {
	inMemory := strings.Contains(rawURL, ":memory:")

	u.Scheme = "file"
	q := u.Query()
	q.Set("_pragma", "foreign_keys(0)")
	u.RawQuery = q.Encode()
	connURL := strings.ReplaceAll(u.String(), "file://", "file:")

	db, err := sql.Open("sqlite", connURL)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if hook != nil {
		hook(db)
	}

	source := &sqlmigrate.EmbedFileSystemMigrationSource{
		FileSystem: sqlite.MigrationsFS,
		Root:       "migrations",
	}

	if inMemory {
		seedMigrationsIfNeeded(ctx, db, "sqlite3")
		n, err := sqlmigrate.Exec(db, "sqlite3", source, sqlmigrate.Up)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("run migrations: %w", err)
		}
		if n > 0 {
			slog.Info("applied sqlite migrations", "count", n)
		}
		if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
			db.Close()
			return nil, fmt.Errorf("enable foreign keys: %w", err)
		}
		return sqlite.NewStore(db), nil
	}

	seedMigrationsIfNeeded(ctx, db, "sqlite3")
	n, err := sqlmigrate.Exec(db, "sqlite3", source, sqlmigrate.Up)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	if n > 0 {
		slog.Info("applied sqlite migrations", "count", n)
	}
	db.Close()

	q.Set("_pragma", "foreign_keys(1)")
	u.RawQuery = q.Encode()
	connURL = strings.ReplaceAll(u.String(), "file://", "file:")
	runtimeDB, err := sql.Open("sqlite", connURL)
	if err != nil {
		return nil, fmt.Errorf("re-open sqlite: %w", err)
	}
	if hook != nil {
		hook(runtimeDB)
	}

	return sqlite.NewStore(runtimeDB), nil
}

func openPostgres(ctx context.Context, u *url.URL, hook func(db *sql.DB)) (dbo.Store, error) {
	connStr := u.String()

	migrateDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("open postgres (migration): %w", err)
	}
	defer migrateDB.Close()

	if hook != nil {
		hook(migrateDB)
	}

	source := &sqlmigrate.EmbedFileSystemMigrationSource{
		FileSystem: postgres.MigrationsFS,
		Root:       "migrations",
	}

	seedMigrationsIfNeeded(ctx, migrateDB, "postgres")

	n, err := sqlmigrate.Exec(migrateDB, "postgres", source, sqlmigrate.Up)
	if err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	if n > 0 {
		slog.Info("applied postgres migrations", "count", n)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("open pgxpool: %w", err)
	}

	return postgres.NewStore(pool), nil
}

func seedMigrationsIfNeeded(ctx context.Context, db *sql.DB, driver string) {
	var hasGorp bool
	switch driver {
	case "sqlite3":
		db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='gorp_migrations')",
		).Scan(&hasGorp)
	case "postgres":
		db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name='gorp_migrations')",
		).Scan(&hasGorp)
	}
	if hasGorp {
		return
	}

	var hasToken bool
	switch driver {
	case "sqlite3":
		db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='token')",
		).Scan(&hasToken)
	case "postgres":
		db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name='token')",
		).Scan(&hasToken)
	}
	if !hasToken {
		return
	}

	slog.Info("detected legacy database, seeding migration table")
	switch driver {
	case "sqlite3":
		db.ExecContext(ctx, "CREATE TABLE gorp_migrations (id TEXT NOT NULL PRIMARY KEY, applied_at DATETIME)")
	case "postgres":
		db.ExecContext(ctx, "CREATE TABLE gorp_migrations (id TEXT NOT NULL PRIMARY KEY, applied_at TIMESTAMPTZ)")
	}

	now := time.Now()
	for _, id := range []string{"001_init.sql", "002_add_host.sql"} {
		db.ExecContext(ctx, "INSERT INTO gorp_migrations (id, applied_at) VALUES ($1, $2)", id, now)
	}
	slog.Info("seeded migration table for legacy database")
}
