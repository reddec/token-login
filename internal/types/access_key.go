package types

import (
	"crypto/sha3"
	"crypto/subtle"
	"fmt"

	"github.com/gobwas/glob"
)

func NewAccessKey(hash []byte, hosts, paths []string) (*AccessKey, error) {
	if len(hosts) == 0 {
		hosts = []string{"**"}
	}
	hostGlobs := make([]glob.Glob, len(hosts))
	for i, h := range hosts {
		g, err := glob.Compile(h, '.')
		if err != nil {
			return nil, fmt.Errorf("compile host glob %q: %w", h, err)
		}
		hostGlobs[i] = g
	}

	if len(paths) == 0 {
		paths = []string{"/**"}
	}
	pathGlobs := make([]glob.Glob, len(paths))
	for i, p := range paths {
		g, err := glob.Compile(p, '/')
		if err != nil {
			return nil, fmt.Errorf("compile path glob %q: %w", p, err)
		}
		pathGlobs[i] = g
	}

	return &AccessKey{
		hash:      hash,
		hostGlobs: hostGlobs,
		pathGlobs: pathGlobs,
	}, nil
}

type AccessKey struct {
	hash      []byte
	hostGlobs []glob.Glob
	pathGlobs []glob.Glob
}

func (t *AccessKey) Valid(host, path string, payload []byte) bool {
	if path == "" {
		path = "/"
	}
	// Any host glob matches?
	hostMatch := false
	for _, g := range t.hostGlobs {
		if g.Match(host) {
			hostMatch = true
			break
		}
	}
	if !hostMatch {
		return false
	}
	// Any path glob matches?
	pathMatch := false
	for _, g := range t.pathGlobs {
		if g.Match(path) {
			pathMatch = true
			break
		}
	}
	if !pathMatch {
		return false
	}
	hash := sha3.Sum384(payload)
	return subtle.ConstantTimeCompare(hash[:], t.hash) == 1
}
