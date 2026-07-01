package types

import (
	"crypto/rand"
	"crypto/sha3"
	"database/sql/driver"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	keyDataSize = 32 // private part
	KeyIDSize   = 8  // public part
	TokenSize   = KeyIDSize + keyDataSize
)

var (
	ErrKeySize    = errors.New("key size invalid")
	errCannotScan = errors.New("cannot scan into Headers")
	errKeyIDType  = errors.New("key ID type should be text")
)

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Headers []Header

func (headers Headers) With(name, value string) Headers {
	return append(headers, Header{
		Name:  name,
		Value: value,
	})
}

func (headers Headers) Without(name string) Headers {
	ans := make([]Header, 0, len(headers))
	for i := range headers {
		if headers[i].Name == name {
			continue
		}
		ans = append(ans, headers[i])
	}
	return ans
}

// Scan implements sql.Scanner for JSON headers stored as JSON/JSONB.
func (headers *Headers) Scan(src any) error {
	if src == nil {
		*headers = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into Headers: %w", src, errCannotScan)
	}
	if err := json.Unmarshal(data, headers); err != nil {
		return fmt.Errorf("unmarshal headers: %w", err)
	}
	return nil
}

// Value implements driver.Valuer for JSON headers.
func (headers Headers) Value() (driver.Value, error) {
	if headers == nil {
		return []byte("[]"), nil
	}
	data, err := json.Marshal(headers)
	if err != nil {
		return nil, fmt.Errorf("marshal headers: %w", err)
	}
	return data, nil
}

func NewKey() (Key, error) {
	var key Key
	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return key, fmt.Errorf("read key random data: %w", err)
	}
	return key, nil
}

func ParseKey(value string) (key Key, err error) {
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

func (rt Key) AccessKey(hosts, paths []string) (*AccessKey, error) {
	return NewAccessKey(rt.Hash(), hosts, paths)
}

type KeyID [KeyIDSize]byte

func (kid *KeyID) UnmarshalText(text []byte) error {
	str := string(text)
	data, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(str)
	if err != nil {
		return fmt.Errorf("decode key ID: %w", err)
	}

	if len(data) != KeyIDSize {
		return ErrKeySize
	}
	copy((*kid)[:], data)
	return nil
}

func (kid *KeyID) MarshalText() ([]byte, error) {
	return []byte(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(kid[:])), nil
}

func (kid KeyID) String() string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(kid[:])
}

func (kid *KeyID) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return kid.UnmarshalText([]byte(v))
	case []byte:
		return kid.UnmarshalText(v)
	default:
		return errKeyIDType
	}
}

func (kid KeyID) Value() (driver.Value, error) {
	return kid.String(), nil
}
