package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reddec/token-login/internal/cache"
	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/internal/types"
	"github.com/reddec/token-login/web"
)

// setupToken creates a valid Key, AccessKey, ent.Token, and populates a cache.
// Returns the raw key string (base32) for use in requests and the access log channel.
func setupToken(t *testing.T, host, path string, headers types.Headers, projectSlug string) (*cache.Cache, string, chan web.Hit) {
	t.Helper()

	key, err := types.NewKey()
	require.NoError(t, err)

	hosts := []string{host}
	if host == "" {
		hosts = nil
	}
	paths := []string{path}
	if path == "" {
		paths = nil
	}
	ak, err := types.NewAccessKey(key.Hash(), hosts, paths)
	require.NoError(t, err)

	dbToken := &dbo.Token{
		ID:          1,
		User:        "testuser",
		KeyID:       func() *types.KeyID { k := key.ID(); return &k }(),
		Hosts:       hosts,
		Paths:       paths,
		Headers:     headers,
		ProjectSlug: projectSlug,
	}

	c := cache.New(nil)
	c.Set(cache.State{
		key.ID(): {
			AccessKey: ak,
			DBToken:   dbToken,
		},
	})

	accessLog := make(chan web.Hit, 1)
	return c, key.String(), accessLog
}

func TestAuthHandlerSuccess(t *testing.T) {
	c, rawKey, accessLog := setupToken(t, "", "", nil, "")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test")
	req.Header.Set(web.TokenHeader, rawKey)
	req.Header.Set(web.HostHeader, "example.com")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "testuser", resp.Header.Get(web.AuthUserHeader))
	assert.NotEmpty(t, resp.Header.Get(web.AuthTokenHintHeader))

	// access log should have one entry
	select {
	case hit := <-accessLog:
		assert.Equal(t, int64(1), hit.ID)
	default:
		t.Error("expected hit in access log")
	}
}

func TestAuthHandlerMissingToken(t *testing.T) {
	c, _, accessLog := setupToken(t, "", "", nil, "")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandlerInvalidTokenFormat(t *testing.T) {
	c, _, accessLog := setupToken(t, "", "", nil, "")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test")
	req.Header.Set(web.TokenHeader, "not-a-valid-base32-token")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandlerTokenNotFound(t *testing.T) {
	// Generate a key that's not in the cache
	key, err := types.NewKey()
	require.NoError(t, err)

	c := cache.New(nil)
	accessLog := make(chan web.Hit, 1)

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test")
	req.Header.Set(web.TokenHeader, key.String())

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandlerHostMismatch(t *testing.T) {
	c, rawKey, accessLog := setupToken(t, "*.example.com", "/**", nil, "")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test")
	req.Header.Set(web.TokenHeader, rawKey)
	req.Header.Set(web.HostHeader, "evil.com") // does not match *.example.com

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandlerPathMismatch(t *testing.T) {
	c, rawKey, accessLog := setupToken(t, "", "/api/**", nil, "")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/forbidden/path")
	req.Header.Set(web.TokenHeader, rawKey)
	req.Header.Set(web.HostHeader, "example.com")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandlerTokenFromQueryParam(t *testing.T) {
	c, rawKey, accessLog := setupToken(t, "", "", nil, "")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test?token="+rawKey)
	req.Header.Set(web.HostHeader, "example.com")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestAuthHandlerHeaderOverQueryParam(t *testing.T) {
	// Header token takes priority over query param
	c, rawKey, accessLog := setupToken(t, "", "", nil, "")

	// Create a second (invalid) key string for the query param
	junkKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test?token="+junkKey)
	req.Header.Set(web.TokenHeader, rawKey) // header wins
	req.Header.Set(web.HostHeader, "example.com")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestAuthHandlerDefaultProjectMatch(t *testing.T) {
	// Token has empty project slug (default), no project param in request → match
	c, rawKey, accessLog := setupToken(t, "", "", nil, "")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test")
	req.Header.Set(web.TokenHeader, rawKey)
	req.Header.Set(web.HostHeader, "example.com")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestAuthHandlerExplicitProjectMatch(t *testing.T) {
	// Token has project slug "myapp", request has ?project=myapp → match
	c, rawKey, accessLog := setupToken(t, "", "", nil, "myapp")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test?project=myapp")
	req.Header.Set(web.TokenHeader, rawKey)
	req.Header.Set(web.HostHeader, "example.com")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestAuthHandlerProjectMismatch(t *testing.T) {
	// Token belongs to default project, request specifies ?project=myapp → mismatch
	c, rawKey, accessLog := setupToken(t, "", "", nil, "")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test?project=myapp")
	req.Header.Set(web.TokenHeader, rawKey)
	req.Header.Set(web.HostHeader, "example.com")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandlerCustomHeaders(t *testing.T) {
	headers := types.Headers{
		{Name: "X-Custom", Value: "custom-value"},
		{Name: "X-Another", Value: "another-value"},
	}
	c, rawKey, accessLog := setupToken(t, "", "", headers, "")

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test")
	req.Header.Set(web.TokenHeader, rawKey)
	req.Header.Set(web.HostHeader, "example.com")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "custom-value", resp.Header.Get("X-Custom"))
	assert.Equal(t, "another-value", resp.Header.Get("X-Another"))
}

func TestAuthHandlerAccessLogOverflow(t *testing.T) {
	// Create a channel with buffer size 0 to test non-blocking send
	c, rawKey, _ := setupToken(t, "", "", nil, "")

	accessLog := make(chan web.Hit) // unbuffered — send will fail if no reader

	handler := web.AuthHandler(c, accessLog)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	// This should not block or panic even though no one is reading from accessLog
	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set(web.URLHeader, "/api/test")
	req.Header.Set(web.TokenHeader, rawKey)
	req.Header.Set(web.HostHeader, "example.com")

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
