package server_test

import (
	"context"
	"testing"

	"github.com/reddec/token-login/api"
	"github.com/reddec/token-login/internal/ent"
	"github.com/reddec/token-login/internal/server"
	"github.com/reddec/token-login/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	client, err := ent.New(ctx, "file::memory:?cache=shared", nil)
	require.NoError(t, err)

	defer client.Close()

	aliceCtx := utils.WithUser(ctx, "alice")
	bobCtx := utils.WithUser(ctx, "bob")

	srv := server.New(client)

	secret1, err := srv.CreateToken(aliceCtx, &api.Config{
		Label: api.NewOptString("l1"),
		Host:  api.NewOptString("*.example.com"),
		Path:  api.NewOptString("/**"),
		Headers: []api.NameValue{
			{Name: "foo", Value: "bar"},
			{Name: "x", Value: "y"},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, secret1)
	require.NotEmpty(t, secret1.Key)
	require.NotEmpty(t, secret1.ID)

	secret2, err := srv.CreateToken(aliceCtx, &api.Config{
		Label: api.NewOptString("l2"),
		Host:  api.NewOptString("someparts"),
		Path:  api.NewOptString("/**"),
		Headers: []api.NameValue{
			{Name: "foo", Value: "bar"},
			{Name: "x", Value: "y"},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, secret2)
	require.NotEmpty(t, secret2.Key)
	require.NotEmpty(t, secret2.ID)

	secret3, err := srv.CreateToken(bobCtx, &api.Config{
		Label: api.NewOptString("l3"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, secret3)
	require.NotEmpty(t, secret3.Key)
	require.NotEmpty(t, secret3.ID)

	t.Run("alice can not see bob and vice versa", func(t *testing.T) {
		aliceTokens, err := srv.ListTokens(aliceCtx)
		require.NoError(t, err)
		require.Len(t, aliceTokens, 2)
		require.Equal(t, "l1", aliceTokens[0].Label)
		require.Equal(t, "l2", aliceTokens[1].Label)
		assert.Equal(t, "*.example.com", aliceTokens[0].Host)
		assert.Equal(t, "/**", aliceTokens[0].Path)
		require.Len(t, aliceTokens[0].Headers, 2)
		assert.Equal(t, "foo", aliceTokens[0].Headers[0].Name)
		assert.Equal(t, "bar", aliceTokens[0].Headers[0].Value)
		assert.Equal(t, "x", aliceTokens[0].Headers[1].Name)
		assert.Equal(t, "y", aliceTokens[0].Headers[1].Value)

		bobTokens, err := srv.ListTokens(bobCtx)
		require.NoError(t, err)
		require.Len(t, bobTokens, 1)
		require.Equal(t, "l3", bobTokens[0].Label)
	})

	t.Run("new token keeps key ID and changes secret", func(t *testing.T) {
		newSecret, err := srv.RefreshToken(bobCtx, api.RefreshTokenParams{Token: secret3.ID})
		require.NoError(t, err)
		require.NotEmpty(t, newSecret)
		require.NotEqual(t, secret3.Key, newSecret.Key)
		require.Equal(t, secret3.ID, newSecret.ID)
	})

	t.Run("can not change token for someone else", func(t *testing.T) {
		_, err := srv.RefreshToken(bobCtx, api.RefreshTokenParams{Token: secret2.ID})
		require.Error(t, err)
	})
}
