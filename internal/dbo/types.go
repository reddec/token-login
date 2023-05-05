package dbo

import (
	"database/sql/driver"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
)

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Headers []Header

func (headers *Headers) Scan(value any) error {
	if value == nil {
		return nil
	}
	data, err := toBytes(value)
	if err != nil {
		return fmt.Errorf("convert data to bytes: %w", err)
	}

	if err := json.Unmarshal(data, headers); err != nil {
		return fmt.Errorf("unmarshal headers: %w", err)
	}
	return nil
}

func (headers Headers) Value() (driver.Value, error) {
	if len(headers) == 0 {
		return nil, nil
	}
	return json.Marshal(headers)
}

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

type KeyID [keyIDSize]byte

func (kid KeyID) String() string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(kid[:])
}

func (kid *KeyID) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("key ID type should be text")
	}

	data, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(str)
	if err != nil {
		return fmt.Errorf("decode key ID: %w", err)
	}

	if len(data) != keyIDSize {
		return ErrKeySize
	}
	copy((*kid)[:], data)
	return nil
}

func (kid KeyID) Value() (driver.Value, error) {
	return kid.String(), nil
}

func toBytes(value any) ([]byte, error) {
	if bytes, ok := value.([]byte); ok {
		return bytes, nil
	}
	if str, ok := value.(string); ok {
		return []byte(str), nil
	}
	return nil, errors.New("get bytes possible only from bytes or string")
}
