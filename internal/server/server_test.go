package server_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reddec/token-login/api"
	"github.com/reddec/token-login/internal/ent"
	"github.com/reddec/token-login/internal/server"
	"github.com/reddec/token-login/internal/utils"
)

// defaultProjectFor returns the ID of the user's default project (slug ""),
// creating one if it doesn't exist yet.
func defaultProjectFor(t *testing.T, srv *server.Server, ctx context.Context) int {
	t.Helper()
	list, err := srv.ListProjects(ctx)
	require.NoError(t, err)
	for _, p := range list {
		if p.Slug == "" {
			return p.ID
		}
	}
	p, err := srv.CreateProject(ctx, &api.ProjectConfig{
		Slug:        "",
		Description: api.NewOptString("Default project"),
	})
	require.NoError(t, err)
	return p.ID
}

func TestNew(t *testing.T) {
	ctx := context.Background()
	client, err := ent.New(ctx, "file::memory:?cache=shared", nil)
	require.NoError(t, err)

	defer client.Close()

	aliceCtx := utils.WithUser(ctx, "alice")
	bobCtx := utils.WithUser(ctx, "bob")

	srv := server.New(client)

	// Create per-user default projects
	aliceDefault := defaultProjectFor(t, srv, aliceCtx)
	bobDefault := defaultProjectFor(t, srv, bobCtx)

	secret1, err := srv.CreateToken(aliceCtx, &api.TokenConfig{
		Label:     api.NewOptString("l1"),
		Host:      api.NewOptString("*.example.com"),
		Path:      api.NewOptString("/**"),
		ProjectId: aliceDefault,
		Headers: []api.NameValue{
			{Name: "foo", Value: "bar"},
			{Name: "x", Value: "y"},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, secret1)
	require.NotEmpty(t, secret1.Key)
	require.NotEmpty(t, secret1.ID)

	secret2, err := srv.CreateToken(aliceCtx, &api.TokenConfig{
		Label:     api.NewOptString("l2"),
		Host:      api.NewOptString("someparts"),
		Path:      api.NewOptString("/**"),
		ProjectId: aliceDefault,
		Headers: []api.NameValue{
			{Name: "foo", Value: "bar"},
			{Name: "x", Value: "y"},
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, secret2)
	require.NotEmpty(t, secret2.Key)
	require.NotEmpty(t, secret2.ID)

	secret3, err := srv.CreateToken(bobCtx, &api.TokenConfig{
		Label:     api.NewOptString("l3"),
		ProjectId: bobDefault,
	})
	require.NoError(t, err)
	require.NotEmpty(t, secret3)
	require.NotEmpty(t, secret3.Key)
	require.NotEmpty(t, secret3.ID)

	t.Run("alice can not see bob and vice versa", func(t *testing.T) {
		aliceTokens, err := srv.ListTokens(aliceCtx, api.ListTokensParams{})
		require.NoError(t, err)
		require.Len(t, aliceTokens, 2)
		require.Equal(t, "l1", aliceTokens[1].Label)
		require.Equal(t, "l2", aliceTokens[0].Label)
		assert.Equal(t, "*.example.com", aliceTokens[1].Host)
		assert.Equal(t, "/**", aliceTokens[0].Path)
		require.Len(t, aliceTokens[0].Headers, 2)
		assert.Equal(t, "foo", aliceTokens[0].Headers[0].Name)
		assert.Equal(t, "bar", aliceTokens[0].Headers[0].Value)
		assert.Equal(t, "x", aliceTokens[0].Headers[1].Name)
		assert.Equal(t, "y", aliceTokens[0].Headers[1].Value)

		bobTokens, err := srv.ListTokens(bobCtx, api.ListTokensParams{})
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

	t.Run("delete token", func(t *testing.T) {
		err := srv.DeleteToken(aliceCtx, api.DeleteTokenParams{Token: secret2.ID})
		require.NoError(t, err)

		// verify it's gone
		_, err = srv.GetToken(aliceCtx, api.GetTokenParams{Token: secret2.ID})
		require.Error(t, err)
	})

	t.Run("delete non-existent token", func(t *testing.T) {
		err := srv.DeleteToken(aliceCtx, api.DeleteTokenParams{Token: 99999})
		require.NoError(t, err) // deleting non-existent is not an error
	})

	t.Run("update token label", func(t *testing.T) {
		err := srv.UpdateToken(aliceCtx, &api.TokenPatch{
			Label: api.NewOptString("updated-label"),
		}, api.UpdateTokenParams{Token: secret1.ID})
		require.NoError(t, err)

		tok, err := srv.GetToken(aliceCtx, api.GetTokenParams{Token: secret1.ID})
		require.NoError(t, err)
		assert.Equal(t, "updated-label", tok.Label)
	})

	t.Run("update token host and path", func(t *testing.T) {
		err := srv.UpdateToken(aliceCtx, &api.TokenPatch{
			Host: api.NewOptString("new.example.com"),
			Path: api.NewOptString("/api/**"),
		}, api.UpdateTokenParams{Token: secret1.ID})
		require.NoError(t, err)

		tok, err := srv.GetToken(aliceCtx, api.GetTokenParams{Token: secret1.ID})
		require.NoError(t, err)
		assert.Equal(t, "new.example.com", tok.Host)
		assert.Equal(t, "/api/**", tok.Path)
	})

	t.Run("update token headers", func(t *testing.T) {
		err := srv.UpdateToken(aliceCtx, &api.TokenPatch{
			Headers: []api.NameValue{
				{Name: "X-New", Value: "new-val"},
			},
		}, api.UpdateTokenParams{Token: secret1.ID})
		require.NoError(t, err)

		tok, err := srv.GetToken(aliceCtx, api.GetTokenParams{Token: secret1.ID})
		require.NoError(t, err)
		require.Len(t, tok.Headers, 1)
		assert.Equal(t, "X-New", tok.Headers[0].Name)
		assert.Equal(t, "new-val", tok.Headers[0].Value)
	})

	t.Run("update non-existent token", func(t *testing.T) {
		err := srv.UpdateToken(aliceCtx, &api.TokenPatch{
			Label: api.NewOptString("nope"),
		}, api.UpdateTokenParams{Token: 99999})
		require.Error(t, err)
	})

	t.Run("on update callback", func(t *testing.T) {
		var updatedID int
		srv.OnUpdate(func(id int) {
			updatedID = id
		})

		cred, err := srv.CreateToken(aliceCtx, &api.TokenConfig{
			Label:     api.NewOptString("callback-test"),
			ProjectId: aliceDefault,
		})
		require.NoError(t, err)
		assert.Equal(t, cred.ID, updatedID)
	})

	t.Run("on remove callback", func(t *testing.T) {
		var removedID int
		srv.OnRemove(func(id int) {
			removedID = id
		})

		cred, err := srv.CreateToken(aliceCtx, &api.TokenConfig{
			Label:     api.NewOptString("remove-test"),
			ProjectId: aliceDefault,
		})
		require.NoError(t, err)

		err = srv.DeleteToken(aliceCtx, api.DeleteTokenParams{Token: cred.ID})
		require.NoError(t, err)
		assert.Equal(t, cred.ID, removedID)
	})
}

func TestProjectCRUD(t *testing.T) {
	ctx := context.Background()
	client, err := ent.New(ctx, "file::memory:?cache=shared", nil)
	require.NoError(t, err)
	defer client.Close()

	aliceCtx := utils.WithUser(ctx, "alice")
	srv := server.New(client)

	// Ensure alice has a default project
	_ = defaultProjectFor(t, srv, aliceCtx)

	t.Run("create project", func(t *testing.T) {
		p, err := srv.CreateProject(aliceCtx, &api.ProjectConfig{
			Slug:        "my-project",
			Description: api.NewOptString("A test project"),
		})
		require.NoError(t, err)
		assert.Equal(t, "my-project", p.Slug)
		assert.Equal(t, "A test project", p.Description)
		assert.NotZero(t, p.ID)
		assert.NotZero(t, p.CreatedAt)
	})

	t.Run("list projects includes default", func(t *testing.T) {
		list, err := srv.ListProjects(aliceCtx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 2) // default + my-project
		// default project is always first (ordered by ID)
		assert.Equal(t, "", list[0].Slug)
		assert.Equal(t, "Default project", list[0].Description)
	})

	t.Run("get project", func(t *testing.T) {
		list, err := srv.ListProjects(aliceCtx)
		require.NoError(t, err)
		defaultP := list[0]

		p, err := srv.GetProject(aliceCtx, api.GetProjectParams{Project: defaultP.ID})
		require.NoError(t, err)
		assert.Equal(t, "", p.Slug)
	})

	t.Run("update project", func(t *testing.T) {
		list, err := srv.ListProjects(aliceCtx)
		require.NoError(t, err)
		nonDefault := list[1] // my-project

		err = srv.UpdateProject(aliceCtx, &api.ProjectPatch{
			Description: api.NewOptString("Updated description"),
		}, api.UpdateProjectParams{Project: nonDefault.ID})
		require.NoError(t, err)

		p, err := srv.GetProject(aliceCtx, api.GetProjectParams{Project: nonDefault.ID})
		require.NoError(t, err)
		assert.Equal(t, "my-project", p.Slug) // slug unchanged
		assert.Equal(t, "Updated description", p.Description)
	})

	t.Run("delete project unlinks tokens", func(t *testing.T) {
		list, err := srv.ListProjects(aliceCtx)
		require.NoError(t, err)
		nonDefault := list[1] // my-project

		// create a token in this project
		cred, err := srv.CreateToken(aliceCtx, &api.TokenConfig{
			Label:     api.NewOptString("to-unlink"),
			ProjectId: nonDefault.ID,
		})
		require.NoError(t, err)

		// delete the project
		err = srv.DeleteProject(aliceCtx, api.DeleteProjectParams{Project: nonDefault.ID})
		require.NoError(t, err)

		// verify it's gone
		_, err = srv.GetProject(aliceCtx, api.GetProjectParams{Project: nonDefault.ID})
		require.Error(t, err)

		// verify token was unlinked (projectId = 0, no slug)
		tok, err := srv.GetToken(aliceCtx, api.GetTokenParams{Token: cred.ID})
		require.NoError(t, err)
		assert.Equal(t, 0, tok.ProjectId)
		assert.Equal(t, "", tok.ProjectSlug)
	})

	t.Run("cannot delete default project", func(t *testing.T) {
		list, err := srv.ListProjects(aliceCtx)
		require.NoError(t, err)
		defaultP := list[0]

		err = srv.DeleteProject(aliceCtx, api.DeleteProjectParams{Project: defaultP.ID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete default project")
	})
}

func TestProjectUserIsolation(t *testing.T) {
	ctx := context.Background()
	client, err := ent.New(ctx, "file::memory:?cache=shared", nil)
	require.NoError(t, err)
	defer client.Close()

	aliceCtx := utils.WithUser(ctx, "alice")
	bobCtx := utils.WithUser(ctx, "bob")
	srv := server.New(client)

	// Both users need their own default project
	aliceDefault := defaultProjectFor(t, srv, aliceCtx)
	_ = defaultProjectFor(t, srv, bobCtx)

	// Alice creates a project
	aliceProject, err := srv.CreateProject(aliceCtx, &api.ProjectConfig{Slug: "alice-only"})
	require.NoError(t, err)

	t.Run("bob cannot see alice's project in list", func(t *testing.T) {
		bobList, err := srv.ListProjects(bobCtx)
		require.NoError(t, err)
		for _, p := range bobList {
			assert.NotEqual(t, aliceProject.ID, p.ID, "bob should not see alice's project")
		}
	})

	t.Run("bob cannot get alice's project", func(t *testing.T) {
		_, err := srv.GetProject(bobCtx, api.GetProjectParams{Project: aliceProject.ID})
		require.Error(t, err)
	})

	t.Run("bob cannot update alice's project", func(t *testing.T) {
		err := srv.UpdateProject(bobCtx, &api.ProjectPatch{
			Description: api.NewOptString("hacked"),
		}, api.UpdateProjectParams{Project: aliceProject.ID})
		require.Error(t, err)
	})

	t.Run("bob cannot delete alice's project", func(t *testing.T) {
		err := srv.DeleteProject(bobCtx, api.DeleteProjectParams{Project: aliceProject.ID})
		require.Error(t, err)
	})

	t.Run("bob cannot create token in alice's project", func(t *testing.T) {
		_, err := srv.CreateToken(bobCtx, &api.TokenConfig{
			Label:     api.NewOptString("stolen"),
			ProjectId: aliceProject.ID,
		})
		require.Error(t, err)
	})

	t.Run("bob cannot create token in alice's default project", func(t *testing.T) {
		_, err := srv.CreateToken(bobCtx, &api.TokenConfig{
			Label:     api.NewOptString("stolen-default"),
			ProjectId: aliceDefault,
		})
		require.Error(t, err)
	})

	t.Run("alice can still use her project", func(t *testing.T) {
		p, err := srv.GetProject(aliceCtx, api.GetProjectParams{Project: aliceProject.ID})
		require.NoError(t, err)
		assert.Equal(t, "alice-only", p.Slug)
	})
}

func TestProjectSlugPerUser(t *testing.T) {
	ctx := context.Background()
	client, err := ent.New(ctx, "file::memory:?cache=shared", nil)
	require.NoError(t, err)
	defer client.Close()

	aliceCtx := utils.WithUser(ctx, "alice")
	bobCtx := utils.WithUser(ctx, "bob")
	srv := server.New(client)

	// Create default projects
	_ = defaultProjectFor(t, srv, aliceCtx)
	_ = defaultProjectFor(t, srv, bobCtx)

	t.Run("different users can use same slug", func(t *testing.T) {
		_, err := srv.CreateProject(aliceCtx, &api.ProjectConfig{Slug: "shared-name"})
		require.NoError(t, err)

		_, err = srv.CreateProject(bobCtx, &api.ProjectConfig{Slug: "shared-name"})
		require.NoError(t, err, "different users should be able to use same slug")
	})

	t.Run("same user cannot duplicate slug", func(t *testing.T) {
		_, err := srv.CreateProject(aliceCtx, &api.ProjectConfig{Slug: "unique-for-alice"})
		require.NoError(t, err)

		_, err = srv.CreateProject(aliceCtx, &api.ProjectConfig{Slug: "unique-for-alice"})
		require.Error(t, err, "same user should not be able to duplicate slug")
	})
}

func TestTokenAssignedToProject(t *testing.T) {
	ctx := context.Background()
	client, err := ent.New(ctx, "file::memory:?cache=shared", nil)
	require.NoError(t, err)
	defer client.Close()

	userCtx := utils.WithUser(ctx, "tester")
	srv := server.New(client)
	defaultID := defaultProjectFor(t, srv, userCtx)

	t.Run("token response includes project fields", func(t *testing.T) {
		cred, err := srv.CreateToken(userCtx, &api.TokenConfig{
			Label:     api.NewOptString("test"),
			ProjectId: defaultID,
		})
		require.NoError(t, err)

		tok, err := srv.GetToken(userCtx, api.GetTokenParams{Token: cred.ID})
		require.NoError(t, err)
		assert.Equal(t, defaultID, tok.ProjectId)
		assert.Equal(t, "", tok.ProjectSlug)
	})

	t.Run("token creation with another user's project fails", func(t *testing.T) {
		otherCtx := utils.WithUser(ctx, "other")
		otherDefault := defaultProjectFor(t, srv, otherCtx)

		_, err := srv.CreateToken(userCtx, &api.TokenConfig{
			Label:     api.NewOptString("test"),
			ProjectId: otherDefault,
		})
		require.Error(t, err, "should not be able to create token in another user's project")
	})

	t.Run("token creation with non-existent project fails", func(t *testing.T) {
		_, err := srv.CreateToken(userCtx, &api.TokenConfig{
			Label:     api.NewOptString("test"),
			ProjectId: 99999,
		})
		require.Error(t, err)
	})
}

func TestLastAccessAtSerializationBug(t *testing.T) {
	ctx := context.Background()
	client, err := ent.New(ctx, "file::memory:?cache=shared", nil)
	require.NoError(t, err)
	defer client.Close()

	userCtx := utils.WithUser(ctx, "tester")
	srv := server.New(client)
	defaultID := defaultProjectFor(t, srv, userCtx)

	// Create a token. The schema defaults LastAccessAt to time.Now(),
	// so the token starts with a non-zero LastAccessAt.
	cred, err := srv.CreateToken(userCtx, &api.TokenConfig{
		Label:     api.NewOptString("test"),
		ProjectId: defaultID,
	})
	require.NoError(t, err)

	tok, err := srv.GetToken(userCtx, api.GetTokenParams{Token: cred.ID})
	require.NoError(t, err)

	// Verify the underlying value is non-zero (token was just created)
	require.False(t, tok.LastAccessAt.Value.IsZero(),
		"expected LastAccessAt to be non-zero after creation")

	// Set must be true when the value is present (non-zero)
	require.True(t, tok.LastAccessAt.Set,
		"Set must be true when LastAccessAt has a value")

	// lastAccessAt must appear in JSON when the token has a non-zero value
	data, err := json.Marshal(tok)
	require.NoError(t, err)

	hasField := strings.Contains(string(data), `"lastAccessAt"`)
	require.True(t, hasField,
		"lastAccessAt field must be present in JSON when value is non-zero")
}
