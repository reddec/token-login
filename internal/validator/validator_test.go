package validator_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/internal/dbo/sqllite"
	"github.com/reddec/token-login/internal/validator"
)

func TestValidator_Valid(t *testing.T) {
	ctx := context.Background()
	store, err := sqllite.New("file::memory:?cache=shared", nil)
	require.NoError(t, err)

	v := validator.NewValidator(store, 10, 1*time.Minute)

	key, err := dbo.NewKey()
	require.NoError(t, err)
	require.NoError(t, store.CreateToken(ctx, dbo.TokenParams{
		User: "admin",
		Config: dbo.TokenConfig{
			Label: "demo",
			Path:  "/**",
		},
		Key: key,
	}))

	key2, err := dbo.NewKey()
	require.NoError(t, err)
	require.NoError(t, store.CreateToken(ctx, dbo.TokenParams{
		User: "user",
		Config: dbo.TokenConfig{
			Label: "demo 2",
			Path:  "/hello",
		},
		Key: key2,
	}))

	t.Logf("KeyID 1: %s\nKeyID 2: %s", key.ID().String(), key2.ID().String())

	t.Run("basic test is ok", func(t *testing.T) {
		found, err := v.Valid(ctx, "/", key.String())
		require.NoError(t, err)
		assert.Equal(t, int64(1), found.ID)

		found2, err := v.Valid(ctx, "/hello", key2.String())
		require.NoError(t, err)
		assert.Equal(t, int64(2), found2.ID)
	})

	t.Run("path validation for glob", func(t *testing.T) {
		found, err := v.Valid(ctx, "/something", key.String())
		require.NoError(t, err)
		assert.Equal(t, int64(1), found.ID)
	})

	t.Run("path validation restricted", func(t *testing.T) {
		_, err := v.Valid(ctx, "/something", key2.String())
		require.Error(t, err)
	})

	t.Run("dump stats", func(t *testing.T) {
		s, err := store.FindToken(ctx, key.ID())
		require.NoError(t, err)
		assert.Empty(t, s.Requests)

		_, err = v.Valid(ctx, "/something", key.String())
		require.NoError(t, err)

		err = v.UpdateStats(ctx)
		require.NoError(t, err)

		s, err = store.FindToken(ctx, key.ID())
		require.NoError(t, err)
		assert.NotEmpty(t, s.Requests)
	})
}
