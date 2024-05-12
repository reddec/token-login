package types

import (
	"crypto/subtle"
	"fmt"

	"github.com/gobwas/glob"
	"golang.org/x/crypto/sha3"
)

func NewAccessKey(hash []byte, host, path string) (*AccessKey, error) {
	if host == "" {
		host = "**"
	}
	hostGlob, err := glob.Compile(host, '.')
	if err != nil {
		return nil, fmt.Errorf("compile host glob: %w", err)
	}

	if path == "" {
		path = "/**"
	}
	pathGlob, err := glob.Compile(path, '/')
	if err != nil {
		return nil, fmt.Errorf("compile path glob: %w", err)
	}

	return &AccessKey{
		hash:     hash,
		pathGlob: pathGlob,
		hostGlob: hostGlob,
	}, nil
}

type AccessKey struct {
	hash     []byte
	pathGlob glob.Glob
	hostGlob glob.Glob
}

func (t *AccessKey) Valid(host, path string, payload []byte) bool {
	if !t.hostGlob.Match(host) {
		return false
	}

	if !t.pathGlob.Match(path) {
		return false
	}

	hash := sha3.Sum384(payload)
	return subtle.ConstantTimeCompare(hash[:], t.hash) == 1
}
