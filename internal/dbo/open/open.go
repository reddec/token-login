// Package open provides the database initialization factory.
// It parses the DSN scheme and dispatches to the engine-specific Open
// functions in the sqlite and postgres packages.
package open

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/internal/dbo/postgres"
	"github.com/reddec/token-login/internal/dbo/sqlite"
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
		// url.Parse rejects non-standard authorities like ":memory:"
		// (treated as invalid port). Extract the scheme manually.
		if !strings.Contains(rawURL, ":memory:") {
			return nil, fmt.Errorf("parse DSN: %w", err)
		}
		before, _, _ := strings.Cut(rawURL, "://")
		if before == "" {
			return nil, fmt.Errorf("parse DSN: %w", err)
		}
		switch before {
		case "sqlite", "sqlite3", "file":
			return sqlite.Open(ctx, rawURL, hook) //nolint:wrapcheck
		default:
			return nil, fmt.Errorf("unsupported database scheme %s: %w", before, errUnsupportedScheme)
		}
	}
	switch u.Scheme {
	case "sqlite", "sqlite3", "file":
		return sqlite.Open(ctx, rawURL, hook) //nolint:wrapcheck
	case "postgres":
		return postgres.Open(ctx, rawURL, hook) //nolint:wrapcheck
	default:
		return nil, fmt.Errorf("unsupported database scheme %s: %w", u.Scheme, errUnsupportedScheme)
	}
}
