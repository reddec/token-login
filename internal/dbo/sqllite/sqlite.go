package sqllite

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/reddec/gsql"
	migrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite" // driver

	"github.com/reddec/token-login/internal/dbo"
)

//go:embed migrations
var migrations embed.FS

func New(url string, configurator func(db *sqlx.DB)) (*Store, error) {
	db, err := sqlx.Open("sqlite", url)
	if err != nil {
		return nil, fmt.Errorf("open DB: %w", err)
	}
	if configurator != nil {
		configurator(db)
	}
	_, err = migrate.Exec(db.DB, "sqlite3", migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations,
		Root:       "migrations",
	}, migrate.Up)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &Store{
		db: db,
	}, nil
}

type Store struct {
	db *sqlx.DB
}

func (st *Store) Close() error {
	return st.db.Close()
}

func (st *Store) ListTokens(ctx context.Context, user string) ([]*dbo.Token, error) {
	return gsql.List[*dbo.Token](ctx, st.db, `SELECT * FROM token WHERE "user" = ? ORDER BY id`, user)
}

func (st *Store) CreateToken(ctx context.Context, params dbo.TokenParams) error {
	const q = `
INSERT INTO token (key_id, hash, "user", label, host, path, headers)
VALUES (?, ?, ?, ?, ?, ?, ?)
`
	_, err := st.db.ExecContext(ctx, q, params.Key.ID(), params.Key.Hash(), params.User, params.Config.Label, params.Config.Host, params.Config.Path, params.Config.Headers)
	return err
}

func (st *Store) GetToken(ctx context.Context, ref dbo.TokenRef) (*dbo.Token, error) {
	v, err := gsql.Get[dbo.Token](ctx, st.db, `SELECT * FROM token WHERE id = ? AND "user" = ? LIMIT 1`, ref.ID, ref.User)
	return &v, err
}

func (st *Store) FindToken(ctx context.Context, id dbo.KeyID) (*dbo.Token, error) {
	v, err := gsql.Get[dbo.Token](ctx, st.db, `SELECT * FROM token WHERE key_id = ? LIMIT 1`, id)
	return &v, err
}

func (st *Store) DeleteToken(ctx context.Context, ref dbo.TokenRef) error {
	_, err := st.db.ExecContext(ctx, `DELETE FROM token WHERE id = ? AND "user" = ?`, ref.ID, ref.User)
	return err
}

func (st *Store) UpdateTokenKey(ctx context.Context, ref dbo.TokenRef, key dbo.Key) error {
	_, err := st.db.ExecContext(ctx,
		`UPDATE token SET key_id = ?, hash = ? WHERE id = ? AND "user" = ?`,
		key.ID(), key.Hash(), ref.ID, ref.User)
	return err
}

func (st *Store) UpdateTokenConfig(ctx context.Context, ref dbo.TokenRef, config dbo.TokenConfig) error {
	_, err := st.db.ExecContext(ctx,
		`UPDATE token SET label = ?, path = ?, headers = ?, host = ? WHERE id = ? AND "user" = ?`,
		config.Label, config.Path, config.Headers, config.Host, ref.ID, ref.User)
	return err
}

func (st *Store) UpdateTokensStats(ctx context.Context, stats []dbo.TokenStat) error {
	const q = `
UPDATE token 
SET requests = requests + ?, last_access_at = max(last_access_at, ?) 
WHERE id = ?
`
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	for _, stat := range stats {
		_, err := tx.ExecContext(ctx, q, stat.Hits, stat.Last, stat.Token)
		if err != nil {
			return errors.Join(fmt.Errorf("add stat for %d: %w", stat.Token, err), tx.Rollback())
		}
	}
	return tx.Commit()
}
