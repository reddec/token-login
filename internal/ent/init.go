package ent

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib" // driver
	_ "modernc.org/sqlite"             // driver

	"github.com/reddec/token-login/internal/ent/project"
	"github.com/reddec/token-login/internal/ent/token"
)

func New(ctx context.Context, rawURL string, hook func(db *sql.DB)) (*Client, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}

	client, err := newDBClient(u, hook)
	if err != nil {
		return nil, fmt.Errorf("create db client: %w", err)
	}

	if err := client.Schema.Create(ctx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	if err := migrateData(ctx, client); err != nil {
		slog.Warn("data migration failed (non-fatal, will retry on next start)", "error", err)
	}

	return client, nil
}

func newDBClient(u *url.URL, hook func(db *sql.DB)) (*Client, error) {
	switch u.Scheme {
	case "sqlite", "sqlite3", "file":
		u.Scheme = "file"
		q := u.Query()
		q.Add("_pragma", "foreign_keys(1)")
		u.RawQuery = q.Encode()
		connURL := strings.ReplaceAll(u.String(), "file://", "file:")
		db, err := sql.Open("sqlite", connURL)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}
		if h := hook; h != nil {
			h(db)
		}
		return NewClient(Driver(entsql.OpenDB(dialect.SQLite, db))), nil
	case "postgres":
		db, err := sql.Open("pgx", u.String())
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}
		if h := hook; h != nil {
			h(db)
		}
		return NewClient(Driver(entsql.OpenDB(dialect.Postgres, db))), nil
	default:
		return nil, fmt.Errorf("unknown dialect %s", u.Scheme)
	}
}

// migrateData creates a per-user default project (empty slug) and backfills
// any tokens that don't have a project assigned, grouped by user.
// This is idempotent and safe to run on every startup.
func migrateData(ctx context.Context, client *Client) error {
	unassigned, err := client.Token.Query().
		Where(token.ProjectIDIsNil()).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query unassigned tokens: %w", err)
	}

	byUser := make(map[string][]*Token)
	for _, t := range unassigned {
		byUser[t.User] = append(byUser[t.User], t)
	}

	for user, tokens := range byUser {
		def, err := client.Project.Query().
			Where(project.User(user), project.Slug("")).
			Only(ctx)
		if IsNotFound(err) {
			def, err = client.Project.Create().
				SetSlug("").
				SetUser(user).
				SetDescription("Default project").
				Save(ctx)
			if err != nil {
				return fmt.Errorf("create default project for user %s: %w", user, err)
			}
			slog.Info("created default project for user", "user", user, "id", def.ID)
		} else if err != nil {
			return fmt.Errorf("query default project for user %s: %w", user, err)
		}

		count, err := client.Token.Update().
			Where(token.IDIn(tokensToIDs(tokens)...)).
			SetProjectID(def.ID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("assign tokens to default project for user %s: %w", user, err)
		}
		if count > 0 {
			slog.Info("assigned tokens to default project", "user", user, "count", count, "project_id", def.ID)
		}
	}
	return nil
}

func tokensToIDs(tokens []*Token) []int {
	ids := make([]int, len(tokens))
	for i, t := range tokens {
		ids[i] = t.ID
	}
	return ids
}
