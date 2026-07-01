// Package dbo provides a universal database access layer with engine-specific
// implementations for SQLite and PostgreSQL via sqlc.
//
//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc generate -f sqlc.yaml
package dbo

import (
	"context"
	"io"
	"time"

	"github.com/reddec/token-login/internal/types"
)

// Token is the domain model for an access token.
type Token struct {
	ID           int64         `json:"id"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	KeyID        *types.KeyID  `json:"key_id"`
	Hash         []byte        `json:"-"`
	User         string        `json:"user"`
	Label        string        `json:"label"`
	Paths        []string      `json:"paths"`
	Hosts        []string      `json:"hosts"`
	Headers      types.Headers `json:"headers,omitempty"`
	ProjectID    int64         `json:"project_id"`
	ProjectSlug  string        `json:"project_slug,omitempty"`
	Requests     int64         `json:"requests"`
	LastAccessAt time.Time     `json:"last_access_at"`
}

// Project is the domain model for a project.
type Project struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	User        string    `json:"user"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
}

// StatsEntry holds accumulated request count and last access time.
type StatsEntry struct {
	Hits int64
	Last time.Time
}

// CreateTokenParams contains the fields needed to create a new token.
type CreateTokenParams struct {
	User      string
	Hash      []byte
	KeyID     *types.KeyID
	Label     string
	Hosts     []string
	Paths     []string
	Headers   types.Headers
	ProjectID int64
}

// UpdateTokenParams contains the fields for updating a token's mutable config.
type UpdateTokenParams struct {
	User    string
	ID      int64
	Hosts   *[]string
	Paths   *[]string
	Label   *string
	Headers *types.Headers
}

// CreateProjectParams contains the fields needed to create a new project.
type CreateProjectParams struct {
	User        string
	Slug        string
	Description string
}

// UpdateProjectParams contains the fields for updating a project.
type UpdateProjectParams struct {
	User        string
	ID          int64
	Description string
}

// Store is the universal database access interface.
type Store interface {
	io.Closer

	// Token CRUD — user-scoped for multi-tenant isolation.
	CreateToken(ctx context.Context, p CreateTokenParams) (*Token, error)
	GetToken(ctx context.Context, user string, id int64) (*Token, error)
	GetTokenByID(ctx context.Context, id int64) (*Token, error)
	ListTokens(ctx context.Context, user string, projectID int64) ([]*Token, error)
	UpdateToken(ctx context.Context, p UpdateTokenParams) (int64, error)
	DeleteToken(ctx context.Context, user string, id int64) (int64, error)
	RefreshToken(ctx context.Context, user string, id int64, hash []byte, keyID *types.KeyID) (int64, error)

	// Project CRUD — user-scoped.
	CreateProject(ctx context.Context, p CreateProjectParams) (*Project, error)
	GetProject(ctx context.Context, user string, id int64) (*Project, error)
	ListProjects(ctx context.Context, user string) ([]*Project, error)
	UpdateProject(ctx context.Context, p UpdateProjectParams) (int64, error)
	DeleteProject(ctx context.Context, user string, id int64) ([]int64, error)
	ProjectExists(ctx context.Context, user string, id int64) (bool, error)

	// Cache operations — unfiltered, returns all rows.
	ListAllTokens(ctx context.Context) ([]*Token, error)
	ListAllProjects(ctx context.Context) ([]*Project, error)

	// Stats — transactional batch update.
	UpdateStats(ctx context.Context, stats map[int64]StatsEntry) error
}
