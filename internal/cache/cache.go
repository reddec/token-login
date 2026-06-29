package cache

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/reddec/token-login/internal/ent"
	"github.com/reddec/token-login/internal/ent/token"
	"github.com/reddec/token-login/internal/types"
)

type State map[types.KeyID]*Token

type Token struct {
	AccessKey   *types.AccessKey
	DBToken     *ent.Token
	ProjectSlug string
}

type Cache struct {
	client *ent.Client
	state  struct {
		data State
		lock sync.RWMutex
	}
}

func New(client *ent.Client) *Cache {
	v := &Cache{client: client}
	v.state.data = make(State)
	return v
}

func (v *Cache) Set(state State) {
	v.state.lock.Lock()
	defer v.state.lock.Unlock()
	v.state.data = state
}

// Patch updates state in-place.
// DO NOT use it for mass update - it locks the whole workflow.
func (v *Cache) Patch(kid types.KeyID, key *Token) {
	v.state.lock.Lock()
	defer v.state.lock.Unlock()
	v.state.data[kid] = key
}

func (v *Cache) Drop(id int) {
	// note: for huge (thousands) keys we may want to create secondary index (O(1)) instead of linear search (O(N))
	v.state.lock.Lock()
	defer v.state.lock.Unlock()
	for k, a := range v.state.data {
		if a.DBToken.ID == id {
			delete(v.state.data, k)
			break
		}
	}
}

func (v *Cache) FindByKey(kid types.KeyID) (*Token, bool) {
	// note: support multiple KID (collisions)
	v.state.lock.RLock()
	defer v.state.lock.RUnlock()
	t, ok := v.state.data[kid]
	return t, ok
}

func (v *Cache) PollKeys(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := v.SyncKeys(ctx); err != nil {
			slog.Error("failed sync keys", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (v *Cache) SyncKeys(ctx context.Context) error {
	all, err := v.client.Token.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("query all tokens: %w", err)
	}

	// Build project ID -> slug lookup
	projects, err := v.client.Project.Query().All(ctx)
	if err != nil {
		return fmt.Errorf("query all projects: %w", err)
	}
	projectSlugs := make(map[int]string, len(projects))
	for _, p := range projects {
		projectSlugs[p.ID] = p.Slug
	}

	state := make(State, len(all))

	for _, t := range all {
		ak, err := types.NewAccessKey(t.Hash, t.Host, t.Path)
		if err != nil {
			slog.Warn("failed to create access key", "id", t.ID, "user", t.User, "error", err)
			continue
		}

		projectSlug := ""
		if t.ProjectID != 0 {
			projectSlug = projectSlugs[t.ProjectID]
		}

		state[*t.KeyID] = &Token{
			AccessKey:   ak,
			DBToken:     t,
			ProjectSlug: projectSlug,
		}
	}

	v.Set(state)
	return nil
}

func (v *Cache) SyncKey(ctx context.Context, id int) error {
	t, err := v.client.Token.Query().WithProject().Where(token.ID(id)).Only(ctx)
	if err != nil {
		return fmt.Errorf("get token %v: %w", id, err)
	}

	aKey, err := types.NewAccessKey(t.Hash, t.Host, t.Path)
	if err != nil {
		return fmt.Errorf("create access key %v: %w", id, err)
	}

	projectSlug := ""
	if t.Edges.Project != nil {
		projectSlug = t.Edges.Project.Slug
	}

	v.Patch(*t.KeyID, &Token{
		AccessKey:   aKey,
		DBToken:     t,
		ProjectSlug: projectSlug,
	})
	return nil
}
