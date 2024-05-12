package validator_test

import (
	"sync/atomic"
	"testing"

	"github.com/reddec/token-login/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/reddec/token-login/internal/validator"
)

func TestValidator_Valid(t *testing.T) {
	v := validator.NewValidator(1024)

	state := make(validator.State)

	demo := genKey("", "/**").Into(state)
	demo2 := genKey("", "/hello").Into(state)
	demo3 := genKey("example.com", "/**").Into(state)
	demo4 := genKey("*.example.com", "/**").Into(state)

	v.Set(state)

	t.Run("basic test is ok", func(t *testing.T) {
		found, err := v.Validate("", "/", demo.String())
		require.NoError(t, err)
		assert.Equal(t, 1, found.ID())

		found2, err := v.Validate("", "/hello", demo2.String())
		require.NoError(t, err)
		assert.Equal(t, 2, found2.ID())
	})

	t.Run("path validation for glob", func(t *testing.T) {
		found, err := v.Validate("", "/something", demo.String())
		require.NoError(t, err)
		assert.Equal(t, 1, found.ID())
	})

	t.Run("path validation restricted", func(t *testing.T) {
		_, err := v.Validate("", "/something", demo2.String())
		require.Error(t, err)
	})

	t.Run("dump stats", func(t *testing.T) {
		_ = readStats(v.AccessLog()) // clear

		_, err := v.Validate("", "/something", demo.String())
		require.NoError(t, err)

		const repeats = 2
		for i := 0; i < repeats; i++ {
			_, err = v.Validate("", "/hello", demo2.String())
			require.NoError(t, err)
		}

		stats := readStats(v.AccessLog())

		require.Len(t, stats, 2)
		assert.Equal(t, 1, stats[demo.AccessKey.ID()])
		assert.Equal(t, repeats, stats[demo2.AccessKey.ID()])
	})

	t.Run("valid host is working", func(t *testing.T) {
		_, err := v.Validate("example.com", "/something", demo3.String())
		require.NoError(t, err)
	})

	t.Run("invalid host is not working", func(t *testing.T) {
		_, err := v.Validate("", "/something", demo3.String())
		require.Error(t, err)
	})

	t.Run("valid wildcard host is working", func(t *testing.T) {
		_, err := v.Validate("some.example.com", "/something", demo4.String())
		require.NoError(t, err)
	})

	t.Run("multi-level wildcard host is not working", func(t *testing.T) {
		_, err := v.Validate("another.some.example.com", "/something", demo4.String())
		require.Error(t, err)
	})

	t.Run("wildcard does not support root level", func(t *testing.T) {
		_, err := v.Validate("example.com", "/something", demo4.String())
		require.Error(t, err)
	})
}

func mustAccessKey(id int, hash []byte, host, path string, headers types.Headers) *validator.AccessKey {
	v, err := validator.NewAccessKey(id, hash, host, path, headers)
	if err != nil {
		panic(err)
	}
	return v
}

func genKey(host, path string, headers ...types.Header) *testKey {
	raw, err := types.NewKey()
	if err != nil {
		panic(err)
	}
	id := idSeq.Add(1)
	accessKey := mustAccessKey(int(id), raw.Hash(), host, path, headers)

	return &testKey{
		Secret:    raw,
		AccessKey: accessKey,
	}
}

type testKey struct {
	Secret    types.Key
	AccessKey *validator.AccessKey
}

func (tk *testKey) String() string {
	return tk.Secret.String()
}

func (tk *testKey) Into(state validator.State) *testKey {
	state[tk.Secret.ID()] = tk.AccessKey
	return tk
}

var idSeq atomic.Int64

func readStats(content <-chan validator.Hit) map[int]int {
	ans := make(map[int]int)
	for {
		select {
		case hit := <-content:
			ans[hit.ID]++
		default:
			return ans
		}
	}
}
