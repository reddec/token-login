package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/internal/types"
)

type store struct {
	pool *pgxpool.Pool
	q    *Queries
}

func NewStore(pool *pgxpool.Pool) dbo.Store {
	return &store{pool: pool, q: New(pool)}
}

func (s *store) Close() error {
	s.pool.Close()
	return nil
}

func (s *store) CreateToken(ctx context.Context, p dbo.CreateTokenParams) (*dbo.Token, error) {
	row, err := s.q.CreateToken(ctx, CreateTokenParams{
		KeyID:     *p.KeyID,
		Hash:      p.Hash,
		User:      p.User,
		Label:     p.Label,
		Path:      p.Path,
		Host:      p.Host,
		Headers:   p.Headers,
		ProjectID: p.ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}
	return tokenToDomain(row.ID, row.CreatedAt, row.UpdatedAt, row.KeyID, row.Hash,
		row.User, row.Label, row.Path, row.Host, row.Headers,
		row.ProjectID, row.Requests, row.LastAccessAt, ""), nil
}

func (s *store) GetToken(ctx context.Context, user string, id int64) (*dbo.Token, error) {
	row, err := s.q.GetToken(ctx, GetTokenParams{User: user, ID: id})
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}
	return tokenViewToDomain(row), nil
}

func (s *store) GetTokenByID(ctx context.Context, id int64) (*dbo.Token, error) {
	row, err := s.q.GetTokenByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get token by id: %w", err)
	}
	return tokenViewToDomain(row), nil
}

func (s *store) ListTokens(ctx context.Context, user string, projectID int64) ([]*dbo.Token, error) {
	if projectID != 0 {
		rows, err := s.q.ListTokensByUserAndProject(ctx, ListTokensByUserAndProjectParams{
			User:      user,
			ProjectID: projectID,
		})
		if err != nil {
			return nil, fmt.Errorf("list tokens by project: %w", err)
		}
		out := make([]*dbo.Token, 0, len(rows))
		for _, r := range rows {
			out = append(out, tokenViewToDomain(r))
		}
		return out, nil
	}
	rows, err := s.q.ListTokens(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("list tokens: %w", err)
	}
	out := make([]*dbo.Token, 0, len(rows))
	for _, r := range rows {
		out = append(out, tokenViewToDomain(r))
	}
	return out, nil
}

func (s *store) UpdateToken(ctx context.Context, p dbo.UpdateTokenParams) (int64, error) {
	current, err := s.q.GetToken(ctx, GetTokenParams{User: p.User, ID: p.ID})
	if err != nil {
		return 0, fmt.Errorf("get token for update: %w", err)
	}
	host := current.Host
	path := current.Path
	label := current.Label
	headers := current.Headers
	if p.Host != nil {
		host = *p.Host
	}
	if p.Path != nil {
		path = *p.Path
	}
	if p.Label != nil {
		label = *p.Label
	}
	if p.Headers != nil {
		headers = *p.Headers
	}
	return s.q.UpdateToken(ctx, UpdateTokenParams{
		Host:    host,
		Path:    path,
		Label:   label,
		Headers: headers,
		User:    p.User,
		ID:      p.ID,
	})
}

func (s *store) DeleteToken(ctx context.Context, user string, id int64) (int64, error) {
	return s.q.DeleteToken(ctx, DeleteTokenParams{User: user, ID: id})
}

func (s *store) RefreshToken(ctx context.Context, user string, id int64, hash []byte, keyID *types.KeyID) (int64, error) {
	return s.q.RefreshToken(ctx, RefreshTokenParams{
		Hash:  hash,
		KeyID: *keyID,
		User:  user,
		ID:    id,
	})
}

func (s *store) CreateProject(ctx context.Context, p dbo.CreateProjectParams) (*dbo.Project, error) {
	row, err := s.q.CreateProject(ctx, CreateProjectParams{
		User:        p.User,
		Slug:        p.Slug,
		Description: p.Description,
	})
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return &dbo.Project{
		ID: row.ID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		User: row.User, Slug: row.Slug, Description: row.Description,
	}, nil
}

func (s *store) GetProject(ctx context.Context, user string, id int64) (*dbo.Project, error) {
	row, err := s.q.GetProject(ctx, GetProjectParams{User: user, ID: id})
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &dbo.Project{
		ID: row.ID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		User: row.User, Slug: row.Slug, Description: row.Description,
	}, nil
}

func (s *store) ListProjects(ctx context.Context, user string) ([]*dbo.Project, error) {
	rows, err := s.q.ListProjects(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	out := make([]*dbo.Project, 0, len(rows))
	for _, r := range rows {
		out = append(out, &dbo.Project{
			ID: r.ID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
			User: r.User, Slug: r.Slug, Description: r.Description,
		})
	}
	return out, nil
}

func (s *store) UpdateProject(ctx context.Context, p dbo.UpdateProjectParams) (int64, error) {
	return s.q.UpdateProject(ctx, UpdateProjectParams{
		Description: p.Description,
		User:        p.User,
		ID:          p.ID,
	})
}

func (s *store) DeleteProject(ctx context.Context, user string, id int64) ([]int64, error) {
	tokenIDs, err := s.q.ListTokenIDsByProject(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list token ids: %w", err)
	}
	if _, err := s.q.DeleteProject(ctx, DeleteProjectParams{User: user, ID: id}); err != nil {
		return nil, fmt.Errorf("delete project: %w", err)
	}
	return tokenIDs, nil
}

func (s *store) ProjectExists(ctx context.Context, user string, id int64) (bool, error) {
	return s.q.ProjectExists(ctx, ProjectExistsParams{User: user, ID: id})
}

func (s *store) EnsureDefaultProject(ctx context.Context, user string) (*dbo.Project, error) {
	row, err := s.q.GetDefaultProject(ctx, user)
	if err == nil {
		return &dbo.Project{
			ID: row.ID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			User: row.User, Slug: row.Slug, Description: row.Description,
		}, nil
	}
	row2, err := s.q.CreateDefaultProject(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create default project: %w", err)
	}
	return &dbo.Project{
		ID: row2.ID, CreatedAt: row2.CreatedAt, UpdatedAt: row2.UpdatedAt,
		User: row2.User, Slug: row2.Slug, Description: row2.Description,
	}, nil
}

func (s *store) ListAllTokens(ctx context.Context) ([]*dbo.Token, error) {
	rows, err := s.q.ListAllTokens(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all tokens: %w", err)
	}
	out := make([]*dbo.Token, 0, len(rows))
	for _, r := range rows {
		out = append(out, tokenViewToDomain(r))
	}
	return out, nil
}

func (s *store) ListAllProjects(ctx context.Context) ([]*dbo.Project, error) {
	rows, err := s.q.ListAllProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all projects: %w", err)
	}
	out := make([]*dbo.Project, 0, len(rows))
	for _, r := range rows {
		out = append(out, &dbo.Project{
			ID: r.ID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
			User: r.User, Slug: r.Slug, Description: r.Description,
		})
	}
	return out, nil
}

func (s *store) UpdateStats(ctx context.Context, stats map[int64]dbo.StatsEntry) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)
	for id, entry := range stats {
		if err := q.UpdateTokenStats(ctx, UpdateTokenStatsParams{
			Requests:     entry.Hits,
			LastAccessAt: entry.Last,
			ID:           id,
		}); err != nil {
			return fmt.Errorf("update stats for %d: %w", id, err)
		}
	}
	return tx.Commit(ctx)
}

func tokenViewToDomain(row TokenView) *dbo.Token {
	kid := row.KeyID
	return &dbo.Token{
		ID: row.ID, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		KeyID: &kid, Hash: row.Hash, User: row.User, Label: row.Label,
		Path: row.Path, Host: row.Host, Headers: row.Headers,
		ProjectID: row.ProjectID, ProjectSlug: row.ProjectSlug,
		Requests: row.Requests, LastAccessAt: row.LastAccessAt,
	}
}

func tokenToDomain(
	id int64, createdAt, updatedAt time.Time, keyID types.KeyID, hash []byte,
	user, label, path, host string, headers types.Headers,
	projectID int64, requests int64, lastAccessAt time.Time,
	projectSlug string,
) *dbo.Token {
	kid := keyID
	return &dbo.Token{
		ID: id, CreatedAt: createdAt, UpdatedAt: updatedAt,
		KeyID: &kid, Hash: hash, User: user, Label: label,
		Path: path, Host: host, Headers: headers,
		ProjectID: projectID, ProjectSlug: projectSlug,
		Requests: requests, LastAccessAt: lastAccessAt,
	}
}
