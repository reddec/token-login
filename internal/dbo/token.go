package dbo

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"golang.org/x/crypto/sha3"

	"github.com/reddec/token-login/internal/utils"
)

var ErrKeySize = errors.New("key size invalid")

const (
	keyIDSize   = 8                       // public part
	keyDataSize = 32                      // private part
	HintChars   = (keyIDSize * 6 / 4) - 1 //nolint:gomnd
)

type Token struct {
	ID           int64     `db:"id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	KeyID        KeyID     `db:"key_id"`
	Hash         []byte    `db:"hash"` // sha-384
	User         string    `db:"user"`
	Label        string    `db:"label"`
	Path         string    `db:"path"`
	Headers      Headers   `db:"headers"`
	Requests     int64     `db:"requests"`
	LastAccessAt time.Time `db:"last_access_at"`
	glob         utils.Cached[glob.Glob]
}

func (t *Token) Hint() string {
	return t.KeyID.String()[:HintChars]
}

func (t *Token) Valid(path string, payload []byte) bool {
	pat, err := t.glob.Get(func() (glob.Glob, error) {
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
	if len(data) != keyIDSize+keyDataSize {
		return key, ErrKeySize
	}
	copy(key[:], data)
	return
}

type Key [keyIDSize + keyDataSize]byte

func (rt Key) ID() KeyID {
	var kid KeyID
	copy(kid[:], rt[:keyIDSize])
	return kid
}

func (rt Key) Payload() []byte {
	return rt[keyIDSize:]
}

func (rt Key) String() string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(rt[:])
}

func (rt Key) Hash() []byte {
	s := sha3.Sum384(rt.Payload())
	return s[:]
}
