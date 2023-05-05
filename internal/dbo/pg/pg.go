package pg

import (
	"context"
	"embed"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib" // driver
	"github.com/jmoiron/sqlx"
	"github.com/reddec/gsql"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/reddec/token-login/internal/dbo"
)

//go:embed migrations
var migrations embed.FS

func New(url string, configurator func(db *sqlx.DB)) (*Store, error) {
	db, err := sqlx.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("open DB: %w", err)
	}
	if configurator != nil {
		configurator(db)
	}
	_, err = migrate.Exec(db.DB, "postgres", migrate.EmbedFileSystemMigrationSource{
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
	return gsql.List[*dbo.Token](ctx, st.db, `SELECT * FROM token WHERE "user" = $1 ORDER BY id`, user)
}

func (st *Store) CreateToken(ctx context.Context, params dbo.TokenParams) error {
	const q = `
INSERT INTO token (key_id, hash, "user", label, path, headers)
VALUES ($1, $2, $3, $4, $5, $6)
`
	_, err := st.db.ExecContext(ctx, q, params.Key.ID(), params.Key.Hash(), params.User, params.Config.Label, params.Config.Path, params.Config.Headers)
	return err
}

func (st *Store) GetToken(ctx context.Context, ref dbo.TokenRef) (*dbo.Token, error) {
	v, err := gsql.Get[dbo.Token](ctx, st.db, `SELECT * FROM token WHERE id = $1 AND "user" = $2 LIMIT 1`, ref.ID, ref.User)
	return &v, err
}

func (st *Store) FindToken(ctx context.Context, id dbo.KeyID) (*dbo.Token, error) {
	v, err := gsql.Get[dbo.Token](ctx, st.db, `SELECT * FROM token WHERE key_id = $1 LIMIT 1`, id)
	return &v, err
}

func (st *Store) DeleteToken(ctx context.Context, ref dbo.TokenRef) error {
	_, err := st.db.ExecContext(ctx, `DELETE FROM token WHERE id = $1 AND "user" = $2`, ref.ID, ref.User)
	return err
}

func (st *Store) UpdateTokenKey(ctx context.Context, ref dbo.TokenRef, key dbo.Key) error {
	_, err := st.db.ExecContext(ctx,
		`UPDATE token SET key_id = $1, hash = $2 WHERE id = $3 AND "user" = $4`,
		key.ID(), key.Hash(), ref.ID, ref.User)
	return err
}

func (st *Store) UpdateTokenConfig(ctx context.Context, ref dbo.TokenRef, config dbo.TokenConfig) error {
	_, err := st.db.ExecContext(ctx,
		`UPDATE token SET label = $1, path = $2, headers = $3 WHERE id = $4 AND "user" = $5`,
		config.Label, config.Path, config.Headers, ref.ID, ref.User)
	return err
}

func (st *Store) UpdateTokensStats(ctx context.Context, stats []dbo.TokenStat) error {
	const q = `
UPDATE token 
SET requests = requests + $1, last_access_at = greatest(last_access_at, $2) 
WHERE id = $3
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
