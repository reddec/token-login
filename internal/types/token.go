package types

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"golang.org/x/crypto/sha3"

	"github.com/reddec/token-login/internal/utils"
)

const (
	keyDataSize = 32                      // private part
	KeyIDSize   = 8                       // public part
	HintChars   = (KeyIDSize * 6 / 4) - 1 //nolint:gomnd
	TokenSize   = KeyIDSize + keyDataSize
)

type Token struct {
	ID           int       `db:"id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	KeyID        KeyID     `db:"key_id"`
	Hash         []byte    `db:"hash"` // sha-384
	User         string    `db:"user"`
	Label        string    `db:"label"`
	Path         string    `db:"path"`
	Host         string    `db:"host"`
	Headers      Headers   `db:"headers"`
	Requests     int64     `db:"requests"`
	LastAccessAt time.Time `db:"last_access_at"`
	pathGlob     utils.Cached[glob.Glob]
	hostGlob     utils.Cached[glob.Glob]
}

func (t *Token) Hint() string {
	return t.KeyID.String()[:HintChars]
}

func (t *Token) Valid(host, path string, payload []byte) bool {
	hostPat, err := t.hostGlob.Get(func() (glob.Glob, error) {
		if t.Host == "" {
			return glob.Compile("**", '.')
		}
		return glob.Compile(t.Host, '.')
	})
	if err != nil || !hostPat.Match(host) {
		return false
	}

	pat, err := t.pathGlob.Get(func() (glob.Glob, error) {
		if t.Path == "" {
			return glob.Compile("/**", '/')
		}
		return glob.Compile(t.Path, '/')
	})
	if err != nil || !pat.Match(path) {
		return false
	}

	hash := sha3.Sum384(payload)
	return subtle.ConstantTimeCompare(hash[:], t.Hash) == 1
}

func NewKey() (Key, error) {
	var key Key
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return key, fmt.Errorf("read key random data: %w", err)
	}
	return key, nil
}

func ParseToken(value string) (key Key, err error) {
	data, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(value))
	if err != nil {
		return key, fmt.Errorf("parse: %w", err)
	}
	if len(data) != TokenSize {
		return key, ErrKeySize
	}
	copy(key[:], data)
	return
}

type Key [TokenSize]byte

func (rt Key) ID() KeyID {
	var kid KeyID
	copy(kid[:], rt[:KeyIDSize])
	return kid
}

func (rt Key) Payload() []byte {
	return rt[KeyIDSize:]
}

func (rt Key) String() string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(rt[:])
}

func (rt Key) Hash() []byte {
	s := sha3.Sum384(rt.Payload())
	return s[:]
}