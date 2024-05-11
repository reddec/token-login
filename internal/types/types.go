package types

import (
	"database/sql/driver"
	"encoding/base32"
	"errors"
	"fmt"
)

var ErrKeySize = errors.New("key size invalid")

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
	var ans = make([]Header, 0, len(headers))
	for i := range headers {
		if headers[i].Name == name {
			continue
		}
		ans = append(ans, headers[i])
	}
	return ans
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
	str, ok := value.(string)
	if !ok {
		return errors.New("key ID type should be text")
	}

	return kid.UnmarshalText([]byte(str))
}

func (kid KeyID) Value() (driver.Value, error) {
	return kid.String(), nil
}
