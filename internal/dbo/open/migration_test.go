package open_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reddec/token-login/internal/dbo"
	"github.com/reddec/token-login/internal/dbo/open"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMigration(t *testing.T) {
	t.Run("v1.0.0/sqlite", func(t *testing.T) {
		testE2EMigration(t, "1.0.0", "sqlite", "")
	})
	t.Run("v1.1.0/sqlite", func(t *testing.T) {
		testE2EMigration(t, "1.1.0", "sqlite", "")
	})
	t.Run("v1.2.0/sqlite", func(t *testing.T) {
		testE2EMigration(t, "1.2.0", "sqlite", "")
	})
	t.Run("v1.0.0/postgres", func(t *testing.T) {
		testE2EMigration(t, "1.0.0", "postgres", pgAddr(t))
	})
	t.Run("v1.1.0/postgres", func(t *testing.T) {
		testE2EMigration(t, "1.1.0", "postgres", pgAddr(t))
	})
	t.Run("v1.2.0/postgres", func(t *testing.T) {
		testE2EMigration(t, "1.2.0", "postgres", pgAddr(t))
	})
}

func pgAddr(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	ctr, err := tcpostgres.Run(ctx,
		"postgres:14",
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Skipf("postgres container unavailable: %v", err)
	}
	t.Cleanup(func() { _ = ctr.Terminate(ctx) })
	host, err := ctr.Host(ctx)
	require.NoError(t, err)
	port, err := ctr.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	return net.JoinHostPort(host, port.Port())
}

func dbURLForContainer(driver, dataDir, pgAddr string) string {
	_ = dataDir
	switch driver {
	case "sqlite":
		return "sqlite:///data/test.db?cache=shared"
	case "postgres":
		// Container can't reach localhost; use host-gateway.
		_, port, _ := net.SplitHostPort(pgAddr)
		return fmt.Sprintf("postgres://test:test@host.docker.internal:%s/testdb?sslmode=disable", port)
	}
	panic("unknown driver: " + driver)
}

func dbURLForHost(t *testing.T, driver, dataDir, pgAddr string) string {
	t.Helper()
	switch driver {
	case "sqlite":
		return "sqlite://" + filepath.Join(dataDir, "test.db")
	case "postgres":
		return fmt.Sprintf("postgres://test:test@%s/testdb?sslmode=disable", pgAddr)
	}
	panic("unknown driver: " + driver)
}

func testE2EMigration(t *testing.T, oldVersion, driver, pgAddr string) {
	t.Helper()
	ctx := context.Background()
	dataDir := t.TempDir()
	containerDSN := dbURLForContainer(driver, dataDir, pgAddr)
	hostDSN := dbURLForHost(t, driver, dataDir, pgAddr)
	oldPort, stopOld := startOldServer(t, ctx, oldVersion, containerDSN, dataDir)
	time.Sleep(3 * time.Second)
	oldBase := fmt.Sprintf("http://localhost:%d", oldPort)
	populateViaOldAPI(t, oldBase, oldVersion)
	require.NoError(t, stopOld())
	store, err := open.Open(ctx, hostDSN, nil)
	require.NoError(t, err)
	defer store.Close()
	verifyAfterMigration(t, store, oldVersion)
}

const testBcrypt = "$2a$10$047XfiR18tkU28qtAFBg8uaKKXd2NOz.w7O8gldzP3Y.k/4KFf6Za"

func startOldServer(t *testing.T, ctx context.Context, version, dbURL, dataDir string) (int, func() error) {
	t.Helper()
	env := map[string]string{"DB_URL": dbURL}
	switch version {
	case "1.0.0":
		env["LOGIN"] = "basic"
		env["BASIC_USER"] = "admin"
		env["BASIC_PASSWORD"] = testBcrypt
	case "1.1.0", "1.2.0":
		env["LOGIN"] = "proxy"
	default:
		t.Fatalf("unsupported old version: %s", version)
	}
	user := fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid())
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("ghcr.io/reddec/token-login:%s", version),
		Env:          env,
		User:         user,
		ExposedPorts: []string{"8080/tcp", "8081/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("8080/tcp"),
			wait.ForListeningPort("8081/tcp"),
		).WithStartupTimeout(30 * time.Second),
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.Binds = []string{dataDir + ":/data"}
			hc.ExtraHosts = []string{"host.docker.internal:host-gateway"}
		},
	}
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "start old server %s", version)
	stop := func() error { return ctr.Terminate(ctx) }
	port, err := ctr.MappedPort(ctx, "8080/tcp")
	require.NoError(t, err)
	return atoi(port.Port()), stop
}

func populateViaOldAPI(t *testing.T, baseURL, version string) {
	t.Helper()
	if version == "1.2.0" {
		populateViaREST(t, baseURL)
		return
	}
	postOld(t, baseURL+"/", version, url.Values{"label": {"minimal"}, "path": {"/**"}})
	postOld(t, baseURL+"/", version, url.Values{"label": {"with path & headers"}, "path": {"/api/**"}})
	if isProxy(version) {
		postOldAsUser(t, baseURL+"/", version, "bob", url.Values{"label": {"bob's token"}, "path": {"/public/**"}})
	} else {
		postOld(t, baseURL+"/", version, url.Values{"label": {"bob's token"}, "path": {"/public/**"}})
	}
	addHeaderOld(t, baseURL, version, 2, "foo", "bar")
	addHeaderOld(t, baseURL, version, 2, "x", "y")
}

// populateViaREST uses the v1.2.0 JSON REST API (/api/v1/tokens).
func populateViaREST(t *testing.T, baseURL string) {
	t.Helper()
	apiBase := baseURL + "/api/v1/tokens"

	// Token 1: minimal.
	restPost(t, apiBase, "admin", `{"label":"minimal","path":"/**"}`)

	// Token 2: with path.
	restPost(t, apiBase, "admin", `{"label":"with path & headers","path":"/api/**"}`)

	// Token 3: bob.
	restPost(t, apiBase, "bob", `{"label":"bob's token","path":"/public/**"}`)

	// Add headers to token 2 via PATCH.
	restPatch(t, apiBase+"/2", "admin", `{"headers":[{"name":"foo","value":"bar"},{"name":"x","value":"y"}]}`)
}

func restPost(t *testing.T, urlStr, user, body string) {
	t.Helper()
	restDo(t, http.MethodPost, urlStr, user, body)
}

func restPatch(t *testing.T, urlStr, user, body string) {
	t.Helper()
	restDo(t, http.MethodPatch, urlStr, user, body) // v1.2.0 may use PATCH or PUT
}

func restDo(t *testing.T, method, urlStr, user, body string) {
	t.Helper()
	req, err := http.NewRequest(method, urlStr, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User", user)
	resp, err := oldClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("REST API %s %s (user=%s): status %d, body: %s",
			method, urlStr, user, resp.StatusCode, respBody)
	}
}

func postOld(t *testing.T, urlStr, version string, form url.Values) {
	postOldAsUser(t, urlStr, version, "admin", form)
}

func postOldAsUser(t *testing.T, urlStr, version, user string, form url.Values) {
	doOldRequest(t, http.MethodPost, urlStr, version, user, form)
}

func addHeaderOld(t *testing.T, baseURL, version string, tokenID int64, name, value string) {
	doOldRequest(t, http.MethodPost,
		fmt.Sprintf("%s/token/%d/", baseURL, tokenID),
		version, "admin",
		url.Values{"name": {name}, "value": {value}, "action": {"headers"}},
	)
}

var oldClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func doOldRequest(t *testing.T, method, urlStr, version, user string, form url.Values) {
	t.Helper()
	req, err := http.NewRequest(method, urlStr, strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	switch version {
	case "1.0.0":
		req.SetBasicAuth(user, "test")
	case "1.1.0":
		req.Header.Set("X-User", user)
	}
	resp, err := oldClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("old API %s %s (v=%s user=%s): status %d, body: %s",
			method, urlStr, version, user, resp.StatusCode, body)
	}
}

func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func verifyAfterMigration(t *testing.T, store dbo.Store, version string) {
	t.Helper()
	ctx := context.Background()
	all, err := store.ListAllTokens(ctx)
	if err != nil {
		t.Logf("ListAllTokens error: %v (type=%T, string=%q)", err, err, err.Error())
	}
	require.NoError(t, err)
	require.Len(t, all, 3, "all tokens should survive migration")
	byLabel := mapTokens(all)
	tk, ok := byLabel["minimal"]
	require.True(t, ok)
	assert.Equal(t, "admin", tk.User)
	assert.Equal(t, []string{"/**"}, tk.Paths)
	assert.NotZero(t, tk.ProjectID)
	tk, ok = byLabel["with path & headers"]
	require.True(t, ok)
	assert.Equal(t, "admin", tk.User)
	assert.Equal(t, []string{"/api/**"}, tk.Paths)
	assert.Len(t, tk.Headers, 2)
	tk, ok = byLabel["bob's token"]
	require.True(t, ok)
	if isProxy(version) {
		assert.Equal(t, "bob", tk.User)
		assert.NotEqual(t, byLabel["minimal"].ProjectID, tk.ProjectID)
		bobTokens, _ := store.ListTokens(ctx, "bob", 0)
		assert.Len(t, bobTokens, 1)
		adminTokens, _ := store.ListTokens(ctx, "admin", 0)
		assert.Len(t, adminTokens, 2)
	}
}

func mapTokens(tokens []*dbo.Token) map[string]*dbo.Token {
	m := make(map[string]*dbo.Token, len(tokens))
	for _, tk := range tokens {
		m[tk.Label] = tk
	}
	return m
}

func isProxy(version string) bool { return version == "1.1.0" || version == "1.2.0" }
