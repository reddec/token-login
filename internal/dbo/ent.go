package dbo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/reddec/token-login/internal/ent"
	"github.com/reddec/token-login/internal/ent/token"
	"github.com/reddec/token-login/internal/types"

	_ "github.com/jackc/pgx/v5/stdlib" // driver
	_ "modernc.org/sqlite"             // driver
)

func New(ctx context.Context, rawURL string, hook func(db *sql.DB)) (*Ent, error) { //nolint:ireturn
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
	return &Ent{dbClient: client}, nil
}

func newDBClient(u *url.URL, hook func(db *sql.DB)) (*ent.Client, error) {
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
		return ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db))), nil
	case "postgres":
		db, err := sql.Open("pgx", u.String())
		if err != nil {
			return nil, err
		}
		if h := hook; h != nil {
			h(db)
		}
		return ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db))), nil
	default:
		return nil, fmt.Errorf("unknown dialect %s", u.Scheme)
	}
}

type Ent struct {
	dbClient *ent.Client
}

func (tmp *Ent) Close() error {
	return tmp.dbClient.Close()
}

func (tmp *Ent) GetToken(ctx context.Context, ref TokenRef) (*types.Token, error) {
	res, err := tmp.dbClient.Token.Query().Where(
		token.ID(ref.ID),
		token.User(ref.User),
	).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	return mapToken(res), nil
}

func (tmp *Ent) FindToken(ctx context.Context, id types.KeyID) (*types.Token, error) {
	res, err := tmp.dbClient.Token.Query().Where(
		token.KeyID(&id),
	).Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("find token: %w", err)
	}
	return mapToken(res), nil
}

func (tmp *Ent) ListTokens(ctx context.Context, user string) ([]*types.Token, error) {
	res, err := tmp.dbClient.Token.Query().Where(
		token.User(user),
	).Order(token.ByID()).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	var ans = make([]*types.Token, 0, len(res))
	for _, r := range res {
		ans = append(ans, mapToken(r))
	}

	return ans, nil
}

func (tmp *Ent) CreateToken(ctx context.Context, params TokenParams) error {
	kid := params.Key.ID()
	return tmp.dbClient.Token.Create().
		SetUser(params.User).
		SetHash(params.Key.Hash()).
		SetKeyID(&kid).
		SetLabel(params.Config.Label).
		SetHeaders(params.Config.Headers).
		SetHost(params.Config.Host).
		SetPath(params.Config.Path).
		Exec(ctx)
}

func (tmp *Ent) DeleteToken(ctx context.Context, ref TokenRef) error {
	_, err := tmp.dbClient.Token.Delete().Where(token.ID(ref.ID), token.User(ref.User)).Exec(ctx)
	return err
}

func (tmp *Ent) UpdateTokenKey(ctx context.Context, ref TokenRef, key types.Key) error {
	kid := key.ID()
	return tmp.dbClient.Token.Update().Where(
		token.User(ref.User),
		token.ID(ref.ID),
	).SetHash(key.Hash()).SetKeyID(&kid).Exec(ctx)
}

func (tmp *Ent) UpdateTokenConfig(ctx context.Context, ref TokenRef, config TokenConfig) error {
	return tmp.dbClient.Token.Update().Where(
		token.User(ref.User),
		token.ID(ref.ID),
	).
		SetLabel(config.Label).
		SetHeaders(config.Headers).
		SetHost(config.Host).
		SetPath(config.Path).
		Exec(ctx)
}

func (tmp *Ent) UpdateTokensStats(ctx context.Context, stats []TokenStat) error {
	tx, err := tmp.dbClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	for _, stat := range stats {
		lastAccess := stat.Last
		// ugly hack due to ent limitations: SET x = max(value, x)
		t, err := tx.Token.Get(ctx, stat.Token)
		if err == nil && t.LastAccessAt.After(lastAccess) {
			lastAccess = t.LastAccessAt
		}
		if err := tx.Token.Update().Where(token.ID(stat.Token)).AddRequests(stat.Hits).SetLastAccessAt(lastAccess).Exec(ctx); err != nil {
			return errors.Join(fmt.Errorf("update stat %v: %w", stat.Token, err), tx.Rollback())
		}
	}

	return tx.Commit()
}

func mapToken(res *ent.Token) *types.Token {
	return &types.Token{
		ID:           res.ID,
		CreatedAt:    res.CreatedAt,
		UpdatedAt:    res.UpdatedAt,
		KeyID:        *res.KeyID,
		Hash:         res.Hash,
		User:         res.User,
		Label:        res.Label,
		Path:         res.Path,
		Host:         res.Host,
		Headers:      res.Headers,
		Requests:     res.Requests,
		LastAccessAt: res.LastAccessAt,
	}
}
