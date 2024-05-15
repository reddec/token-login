package types_test

import (
	"testing"

	"github.com/reddec/token-login/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Valid(t *testing.T) {

	demo := genKey("", "/**")
	demo2 := genKey("", "/hello")
	demo3 := genKey("example.com", "/**")
	demo4 := genKey("*.example.com", "/**")

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
}

func mustAccessKey(hash []byte, host, path string) *types.AccessKey {
	v, err := types.NewAccessKey(hash, host, path)
	if err != nil {
		panic(err)
	}
	return v
}

func genKey(host, path string) *testKey {
	raw, err := types.NewKey()
	if err != nil {
		panic(err)
	}
	accessKey := mustAccessKey(raw.Hash(), host, path)

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
