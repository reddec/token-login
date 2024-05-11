package dbo_test

import (
	"context"
	"testing"
	"time"

	"github.com/reddec/token-login/internal/ent"
	"github.com/reddec/token-login/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reddec/token-login/internal/dbo"
)

func TestInterface_pg(t *testing.T) {
	ctx := context.Background()
	store, err := dbo.New(ctx, "postgres://postgres:postgres@localhost", nil)
	require.NoError(t, err)

	defer store.Close()

	testStore(ctx, t, store)
}

func TestInterface_sqlite(t *testing.T) {
	ctx := context.Background()
	store, err := dbo.New(ctx, "file::memory:?cache=shared", nil)
	require.NoError(t, err)

	defer store.Close()

	testStore(ctx, t, store)
}

func testStore(ctx context.Context, t *testing.T, store dbo.Storage) {
	key1, err := types.NewKey()
	require.NoError(t, err)
	key2, err := types.NewKey()
	require.NoError(t, err)
	key3, err := types.NewKey()
	require.NoError(t, err)

	adminTokenParams := dbo.TokenParams{
		User: "admin",
		Config: dbo.TokenConfig{
			Label: "admin token",
			Path:  "/**",
			Headers: types.Headers{
				{Name: "X-Group", Value: "sysadmin"},
			},
		},
		Key: key1,
	}

	adminTokenParams2 := dbo.TokenParams{
		User: "admin",
		Config: dbo.TokenConfig{
			Label: "admin token 2",
			Path:  "/2",
			Headers: types.Headers{
				{Name: "X-Group", Value: "sysadmin-2"},
			},
		},
		Key: key2,
	}

	guestTokenParams := dbo.TokenParams{
		User: "guest",
		Config: dbo.TokenConfig{
			Label: "guest token",
			Path:  "/guest",
			Headers: types.Headers{
				{Name: "X-Group", Value: "guest"},
			},
		},
		Key: key3,
	}

	err = store.CreateToken(ctx, adminTokenParams)
	require.NoError(t, err)

	err = store.CreateToken(ctx, adminTokenParams2)
	require.NoError(t, err)

	err = store.CreateToken(ctx, guestTokenParams)
	require.NoError(t, err)

	t.Run("list tokens", func(t *testing.T) {
		admins, err := store.ListTokens(ctx, "admin")
		require.NoError(t, err)
		require.Len(t, admins, 2)
		assertParams(t, admins[0], adminTokenParams)
		assertParams(t, admins[1], adminTokenParams2)
	})

	t.Run("find token by key", func(t *testing.T) {
		token, err := store.FindToken(ctx, guestTokenParams.Key.ID())
		require.NoError(t, err)
		assertParams(t, token, guestTokenParams)
	})

	t.Run("get token by ref", func(t *testing.T) {
		s, err := store.FindToken(ctx, guestTokenParams.Key.ID())
		require.NoError(t, err)

		token, err := store.GetToken(ctx, dbo.TokenRef{
			User: guestTokenParams.User,
			ID:   s.ID,
		})
		require.NoError(t, err)
		assertParams(t, token, guestTokenParams)
	})

	t.Run("update token key", func(t *testing.T) {
		s, err := store.FindToken(ctx, guestTokenParams.Key.ID())
		require.NoError(t, err)

		keyNG, err := types.NewKey()
		require.NoError(t, err)

		err = store.UpdateTokenKey(ctx, dbo.TokenRef{
			User: s.User,
			ID:   s.ID,
		}, keyNG)
		require.NoError(t, err)

		_, err = store.FindToken(ctx, guestTokenParams.Key.ID())
		require.Error(t, err)

		sNG, err := store.FindToken(ctx, keyNG.ID())
		assert.Equal(t, sNG.ID, s.ID)

		guestTokenParams.Key = keyNG
	})

	t.Run("update token params", func(t *testing.T) {
		s, err := store.FindToken(ctx, guestTokenParams.Key.ID())
		require.NoError(t, err)

		err = store.UpdateTokenConfig(ctx, dbo.TokenRef{
			User: guestTokenParams.User,
			ID:   s.ID,
		}, dbo.TokenConfig{
			Label: "NG",
			Path:  "/ng",
		})
		require.NoError(t, err)

		sNG, err := store.FindToken(ctx, guestTokenParams.Key.ID())
		require.NoError(t, err)
		assert.Equal(t, "NG", sNG.Label)
		assert.Equal(t, "/ng", sNG.Path)
		assert.Empty(t, sNG.Headers)
	})

	t.Run("update token stats", func(t *testing.T) {
		s, err := store.FindToken(ctx, guestTokenParams.Key.ID())
		require.NoError(t, err)
		assert.Empty(t, s.Requests)
		now := time.Now()
		err = store.UpdateTokensStats(ctx, []dbo.TokenStat{
			{
				Token: s.ID,
				Last:  now,
				Hits:  1,
			},
		})
		require.NoError(t, err)

		sNG, err := store.FindToken(ctx, guestTokenParams.Key.ID())
		require.NoError(t, err)
		assert.Equal(t, int64(1), sNG.Requests)
		assert.True(t, now.Round(time.Millisecond).Equal(sNG.LastAccessAt.Round(time.Millisecond)))
	})

	t.Run("delete token", func(t *testing.T) {
		key, err := types.NewKey()
		require.NoError(t, err)

		err = store.CreateToken(ctx, dbo.TokenParams{
			User: "xxx",
			Key:  key,
		})
		require.NoError(t, err)

		s, err := store.FindToken(ctx, key.ID())
		require.NoError(t, err, "token should be found")

		err = store.DeleteToken(ctx, dbo.TokenRef{
			User: "xxx",
			ID:   s.ID,
		})
		require.NoError(t, err)

		_, err = store.FindToken(ctx, key.ID())
		require.Error(t, err, "token should NOT be found")
		assert.True(t, ent.IsNotFound(err))
	})
}

func assertParams(t *testing.T, token *types.Token, params dbo.TokenParams) {
	assert.Equal(t, params.Key.ID(), token.KeyID)
	assert.Equal(t, params.User, token.User)
	assert.Equal(t, params.Config.Label, token.Label)
	assert.Equal(t, params.Config.Path, token.Path)
	assert.Equal(t, params.Config.Headers, token.Headers)
}
