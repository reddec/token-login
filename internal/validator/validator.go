package validator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/internal/utils"
)

var ErrInvalidToken = errors.New("invalid token")

type Storage interface {
	FindToken(ctx context.Context, id dbo.KeyID) (*dbo.Token, error)
	UpdateTokensStats(ctx context.Context, stats []dbo.TokenStat) error
}

type Validator struct {
	stats utils.Stats
	cache *cachedStorage
	store Storage
}

func NewValidator(storage Storage, cacheCapacity int, cacheTTL time.Duration) *Validator {
	v := &Validator{
		store: storage,
		cache: newCachedStorage(storage, cacheTTL, cacheCapacity),
	}
	return v
}

func (v *Validator) Invalidate(keyID dbo.KeyID) {
	v.cache.Clear(keyID)
}

func (v *Validator) Valid(ctx context.Context, host, path string, token string) (*dbo.Token, error) {
	t, err := v.validate(ctx, host, path, token)
	if err != nil {
		return t, err
	}
	v.stats.Inc(t.ID)
	return t, nil
}

func (v *Validator) validate(ctx context.Context, host, path string, token string) (*dbo.Token, error) {
	// parse token
	key, err := dbo.ParseToken(token)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	dbToken, err := v.cache.Get(ctx, key.ID())
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	if !dbToken.Valid(host, path, key.Payload()) {
		return nil, ErrInvalidToken
	}
	return dbToken, nil
}

func (v *Validator) UpdateStats(ctx context.Context) error {
	stats := v.stats.Pop()
	if len(stats) == 0 {
		return nil
	}

	var storeStats = make([]dbo.TokenStat, 0, len(stats))
	for _, s := range stats {
		storeStats = append(storeStats, dbo.TokenStat{
			Token: s.ID,
			Last:  time.Unix(0, s.Last),
			Hits:  s.Requests,
		})
	}

	return v.store.UpdateTokensStats(ctx, storeStats)
}
