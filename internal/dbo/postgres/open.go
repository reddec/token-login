package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/reddec/token-login/internal/dbo"
	migrate "github.com/rubenv/sql-migrate"
)

// Open opens a PostgreSQL database, runs migrations, and returns a Store.
//
// The URL must use the postgres scheme. The connection string is passed
// directly to pgx.
func Open(ctx context.Context, rawURL string, hook func(db *sql.DB)) (dbo.Store, error) {
	migrateDB, err := sql.Open("pgx", rawURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres (migration): %w", err)
	}
	defer func() { _ = migrateDB.Close() }()

	if hook != nil {
		hook(migrateDB)
	}

	source := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: MigrationsFS,
		Root:       "migrations",
	}

	n, err := migrate.Exec(migrateDB, "postgres", source, migrate.Up)
	if err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	if n > 0 {
		slog.Info("applied postgres migrations", "count", n)
	}

	pool, err := pgxpool.New(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("open pgxpool: %w", err)
	}

	return NewStore(pool), nil
}
