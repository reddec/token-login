package ent

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib" // driver
	_ "modernc.org/sqlite"             // driver
)

func New(ctx context.Context, rawURL string, hook func(db *sql.DB)) (*Client, error) { //nolint:ireturn
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
			return nil, err
		}
		if h := hook; h != nil {
			h(db)
		}
		return NewClient(Driver(entsql.OpenDB(dialect.SQLite, db))), nil
	case "postgres":
		db, err := sql.Open("pgx", u.String())
		if err != nil {
			return nil, err
		}
		if h := hook; h != nil {
			h(db)
		}
		return NewClient(Driver(entsql.OpenDB(dialect.Postgres, db))), nil
	default:
		return nil, fmt.Errorf("unknown dialect %s", u.Scheme)
	}
}
