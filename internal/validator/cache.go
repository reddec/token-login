package validator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/reddec/token-login/internal/types"

	lru "github.com/hashicorp/golang-lru/v2"
)

type cacheItem struct {
	Token   *types.Token
	Created time.Time
	Lock    sync.RWMutex
}

func (ci *cacheItem) expired(ttl time.Duration) bool {
	return time.Since(ci.Created) > ttl
}

func newCachedStorage(store Storage, ttl time.Duration, cacheSize int) *cachedStorage {
	data, err := lru.New[types.KeyID, *cacheItem](cacheSize)
	if err != nil {
		panic(err) // negative cache capacity
	}
	return &cachedStorage{
		store: store,
		ttl:   ttl,
		cache: data,
	}
}

type cachedStorage struct {
	store Storage
	ttl   time.Duration
	cache *lru.Cache[types.KeyID, *cacheItem]
}

func (c *cachedStorage) Clear(id types.KeyID) {
	c.cache.Remove(id)
}

func (c *cachedStorage) Get(ctx context.Context, id types.KeyID) (*types.Token, error) {
	v, ok := c.cache.Get(id)
	if ok {
		v.Lock.RLock()
		token, expired := v.Token, v.expired(c.ttl)
		v.Lock.RUnlock()
		if !expired {
			return token, nil
		}
	}

	// optimistic flow, load can be limited by connections limit
	out, err := c.store.FindToken(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find token: %w", err)
	}

	c.cache.Add(id, &cacheItem{
		Token:   out,
		Created: time.Now(),
	})

	return out, nil
}
