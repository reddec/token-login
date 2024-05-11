package dbo

import (
	"context"
	"io"
	"time"

	"github.com/reddec/token-login/internal/types"
)

type TokenRef struct {
	User string
	ID   int
}

type TokenConfig struct {
	Label   string
	Path    string
	Host    string
	Headers types.Headers
}

type TokenParams struct {
	User   string
	Config TokenConfig
	Key    types.Key
}

type TokenStat struct {
	Token int
	Last  time.Time
	Hits  int64
}

type Storage interface {
	io.Closer
	GetToken(ctx context.Context, ref TokenRef) (*types.Token, error)
	FindToken(ctx context.Context, id types.KeyID) (*types.Token, error)
	ListTokens(ctx context.Context, user string) ([]*types.Token, error)
	CreateToken(ctx context.Context, params TokenParams) error
	DeleteToken(ctx context.Context, ref TokenRef) error
	UpdateTokenKey(ctx context.Context, ref TokenRef, key types.Key) error
	UpdateTokenConfig(ctx context.Context, ref TokenRef, config TokenConfig) error
	UpdateTokensStats(ctx context.Context, stats []TokenStat) error
}
