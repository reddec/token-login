package dbo

import (
	"context"
	"io"
	"time"
)

type TokenRef struct {
	User string
	ID   int64
}

type TokenConfig struct {
	Label   string
	Path    string
	Host    string
	Headers Headers
}

type TokenParams struct {
	User   string
	Config TokenConfig
	Key    Key
}

type TokenStat struct {
	Token int64
	Last  time.Time
	Hits  int64
}

type Storage interface {
	io.Closer
	GetToken(ctx context.Context, ref TokenRef) (*Token, error)
	FindToken(ctx context.Context, id KeyID) (*Token, error)
	ListTokens(ctx context.Context, user string) ([]*Token, error)
	CreateToken(ctx context.Context, params TokenParams) error
	DeleteToken(ctx context.Context, ref TokenRef) error
	UpdateTokenKey(ctx context.Context, ref TokenRef, key Key) error
	UpdateTokenConfig(ctx context.Context, ref TokenRef, config TokenConfig) error
	UpdateTokensStats(ctx context.Context, stats []TokenStat) error
}
