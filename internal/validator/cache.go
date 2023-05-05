package validator

import (
	"context"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/reddec/token-login/internal/dbo"
)

type cacheItem struct {
	Token   *dbo.Token
	Created time.Time
	Lock    sync.RWMutex
}

func (ci *cacheItem) expired(ttl time.Duration) bool {
	return time.Since(ci.Created) > ttl
}

func newCachedStorage(store Storage, ttl time.Duration, cacheSize int) *cachedStorage {
	data, err := lru.New[dbo.KeyID, *cacheItem](cacheSize)
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
	cache *lru.Cache[dbo.KeyID, *cacheItem]
}

func (c *cachedStorage) Clear(id dbo.KeyID) {
	c.cache.Remove(id)
}

func (c *cachedStorage) Get(ctx context.Context, id dbo.KeyID) (*dbo.Token, error) {
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
