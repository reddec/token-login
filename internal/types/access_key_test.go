package types_test

import (
	"testing"

	"github.com/reddec/token-login/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Valid(t *testing.T) {

	demo := genKey([]string{""}, []string{"/**"})
	demo2 := genKey([]string{""}, []string{"/hello"})
	demo3 := genKey([]string{"example.com"}, []string{"/**"})
	demo4 := genKey([]string{"*.example.com"}, []string{"/**"})

	t.Run("basic test is ok", func(t *testing.T) {
		ok := demo.AccessKey.Valid("", "/", demo.Secret.Payload())
		assert.True(t, ok)

		ok2 := demo2.AccessKey.Valid("", "/hello", demo2.Secret.Payload())
		assert.True(t, ok2)
	})

	t.Run("path validation for glob", func(t *testing.T) {
		ok := demo.AccessKey.Valid("", "/something", demo.Secret.Payload())
		assert.True(t, ok)
	})

	t.Run("path validation restricted", func(t *testing.T) {
		ok := demo.AccessKey.Valid("", "/something", demo2.Secret.Payload())
		assert.False(t, ok)
	})

	t.Run("valid host is working", func(t *testing.T) {
		ok := demo3.AccessKey.Valid("example.com", "/something", demo3.Secret.Payload())
		assert.True(t, ok)
	})

	t.Run("invalid host is not working", func(t *testing.T) {
		ok := demo3.AccessKey.Valid("", "/something", demo3.Secret.Payload())
		require.False(t, ok)
	})

	t.Run("valid wildcard host is working", func(t *testing.T) {
		ok := demo4.AccessKey.Valid("some.example.com", "/something", demo4.Secret.Payload())
		require.True(t, ok)
	})

	t.Run("multi-level wildcard host is not working", func(t *testing.T) {
		ok := demo4.AccessKey.Valid("another.some.example.com", "/something", demo4.Secret.Payload())
		require.False(t, ok)
	})

	t.Run("wildcard does not support root level", func(t *testing.T) {
		ok := demo4.AccessKey.Valid("example.com", "/something", demo4.Secret.Payload())
		require.False(t, ok)
	})

	t.Run("multi host globs match any", func(t *testing.T) {
		multi := genKey([]string{"*.example.com", "*.test.com"}, []string{"/**"})
		ok := multi.AccessKey.Valid("foo.example.com", "/anything", multi.Secret.Payload())
		assert.True(t, ok)
		ok = multi.AccessKey.Valid("bar.test.com", "/anything", multi.Secret.Payload())
		assert.True(t, ok)
		ok = multi.AccessKey.Valid("baz.other.com", "/anything", multi.Secret.Payload())
		assert.False(t, ok)
	})

	t.Run("multi path globs match any", func(t *testing.T) {
		multi := genKey([]string{"**"}, []string{"/api/**", "/admin/**"})
		ok := multi.AccessKey.Valid("example.com", "/api/v1/foo", multi.Secret.Payload())
		assert.True(t, ok)
		ok = multi.AccessKey.Valid("example.com", "/admin/users", multi.Secret.Payload())
		assert.True(t, ok)
		ok = multi.AccessKey.Valid("example.com", "/public", multi.Secret.Payload())
		assert.False(t, ok)
	})

	t.Run("empty host list matches any host", func(t *testing.T) {
		empty := genKey(nil, []string{"/**"})
		ok := empty.AccessKey.Valid("any-host.com", "/something", empty.Secret.Payload())
		assert.True(t, ok)
	})

	t.Run("empty path list matches any path", func(t *testing.T) {
		empty := genKey([]string{"**"}, nil)
		ok := empty.AccessKey.Valid("example.com", "/any/path", empty.Secret.Payload())
		assert.True(t, ok)
	})
}

func mustAccessKey(hash []byte, hosts, paths []string) *types.AccessKey {
	v, err := types.NewAccessKey(hash, hosts, paths)
	if err != nil {
		panic(err)
	}
	return v
}

func genKey(hosts, paths []string) *testKey {
	raw, err := types.NewKey()
	if err != nil {
		panic(err)
	}
	accessKey := mustAccessKey(raw.Hash(), hosts, paths)

	return &testKey{
		Secret:    raw,
		AccessKey: accessKey,
	}
}

type testKey struct {
	Secret    types.Key
	AccessKey *types.AccessKey
}

func (tk *testKey) String() string {
	return tk.Secret.String()
}
