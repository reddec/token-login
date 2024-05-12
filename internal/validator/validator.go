package validator

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/reddec/token-login/internal/types"
	"golang.org/x/crypto/sha3"
)

var ErrInvalidToken = errors.New("invalid token")

type State map[types.KeyID]*AccessKey

type Hit struct {
	Time time.Time
	ID   int
}

type Validator struct {
	state struct {
		data State
		lock sync.RWMutex
	}
	accessLog chan Hit
}

func NewValidator(statsBuffer int) *Validator {
	v := &Validator{
		accessLog: make(chan Hit, statsBuffer),
	}
	return v
}

func (v *Validator) Set(state State) {
	v.state.lock.Lock()
	defer v.state.lock.Unlock()
	v.state.data = state
}

func (v *Validator) AccessLog() <-chan Hit {
	return v.accessLog
}

func (v *Validator) Validate(host, path string, token string) (*AccessKey, error) {
	key, err := types.ParseToken(token)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}

	kid := key.ID()

	accessKey, ok := v.findAccessKey(kid)
	if !ok {
		return nil, fmt.Errorf("find key %s: %w", kid.String(), ErrInvalidToken)
	}

	if !accessKey.Valid(host, path, key.Payload()) {
		return nil, fmt.Errorf("validate %s: %w", kid.String(), ErrInvalidToken)
	}

	select {
	case v.accessLog <- Hit{Time: time.Now(), ID: accessKey.ID()}:
	default:
	}

	return accessKey, nil
}

func (v *Validator) findAccessKey(id types.KeyID) (*AccessKey, bool) {
	v.state.lock.RLock()
	defer v.state.lock.RUnlock()
	t, ok := v.state.data[id]
	return t, ok
}

func NewAccessKey(id int, hash []byte, host, path string, headers types.Headers) (*AccessKey, error) {
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
		id:       id,
		hash:     hash,
		headers:  headers,
		pathGlob: pathGlob,
		hostGlob: hostGlob,
	}, nil
}

type AccessKey struct {
	id       int
	hash     []byte
	headers  types.Headers
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

func (t *AccessKey) Headers() types.Headers {
	return t.headers
}

func (t *AccessKey) ID() int {
	return t.id
}
