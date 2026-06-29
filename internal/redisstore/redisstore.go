// Package redisstore implements sessions.Store from oidc-login using Redis.
package redisstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// New creates a new Redis-backed store from a redigo pool.
func New(pool *redis.Pool) *Store {
	return &Store{pool: pool}
}

// Store implements the oidc-login sessions.Store interface using Redis.
type Store struct {
	pool *redis.Pool
}

func (s *Store) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get redis conn: %w", err)
	}
	defer conn.Close()

	seconds := int(ttl.Seconds())
	if seconds <= 0 {
		// delete immediately if expired; this shouldn't happen for Set
		seconds = 1
	}
	_, err = conn.Do("SETEX", key, seconds, value)
	if err != nil {
		return fmt.Errorf("redis SETEX: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, key string) ([]byte, error) {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get redis conn: %w", err)
	}
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil, nil
		}
		return nil, fmt.Errorf("redis GET: %w", err)
	}
	return data, nil
}

func (s *Store) Delete(ctx context.Context, key string) error {
	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get redis conn: %w", err)
	}
	defer conn.Close()

	_, err = conn.Do("DEL", key)
	if err != nil {
		return fmt.Errorf("redis DEL: %w", err)
	}
	return nil
}
