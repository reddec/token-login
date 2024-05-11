// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/reddec/token-login/internal/ent/token"
	"github.com/reddec/token-login/internal/types"
)

// Token is the model entity for the Token schema.
type Token struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// KeyID holds the value of the "key_id" field.
	KeyID *types.KeyID `json:"key_id,omitempty"`
	// Hash holds the value of the "hash" field.
	Hash []byte `json:"hash,omitempty"`
	// User holds the value of the "user" field.
	User string `json:"user,omitempty"`
	// Label holds the value of the "label" field.
	Label string `json:"label,omitempty"`
	// Path holds the value of the "path" field.
	Path string `json:"path,omitempty"`
	// Host holds the value of the "host" field.
	Host string `json:"host,omitempty"`
	// Headers holds the value of the "headers" field.
	Headers types.Headers `json:"headers,omitempty"`
	// Requests holds the value of the "requests" field.
	Requests int64 `json:"requests,omitempty"`
	// LastAccessAt holds the value of the "last_access_at" field.
	LastAccessAt time.Time `json:"last_access_at,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Token) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case token.FieldHash, token.FieldHeaders:
			values[i] = new([]byte)
		case token.FieldID, token.FieldRequests:
			values[i] = new(sql.NullInt64)
		case token.FieldUser, token.FieldLabel, token.FieldPath, token.FieldHost:
			values[i] = new(sql.NullString)
		case token.FieldCreatedAt, token.FieldUpdatedAt, token.FieldLastAccessAt:
			values[i] = new(sql.NullTime)
		case token.FieldKeyID:
			values[i] = token.ValueScanner.KeyID.ScanValue()
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Token fields.
func (t *Token) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case token.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			t.ID = int(value.Int64)
		case token.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				t.CreatedAt = value.Time
			}
		case token.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				t.UpdatedAt = value.Time
			}
		case token.FieldKeyID:
			if value, err := token.ValueScanner.KeyID.FromValue(values[i]); err != nil {
				return err
			} else {
				t.KeyID = value
			}
		case token.FieldHash:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field hash", values[i])
			} else if value != nil {
				t.Hash = *value
			}
		case token.FieldUser:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field user", values[i])
			} else if value.Valid {
				t.User = value.String
			}
		case token.FieldLabel:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field label", values[i])
			} else if value.Valid {
				t.Label = value.String
			}
		case token.FieldPath:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field path", values[i])
			} else if value.Valid {
				t.Path = value.String
			}
		case token.FieldHost:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field host", values[i])
			} else if value.Valid {
				t.Host = value.String
			}
		case token.FieldHeaders:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field headers", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &t.Headers); err != nil {
					return fmt.Errorf("unmarshal field headers: %w", err)
				}
			}
		case token.FieldRequests:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field requests", values[i])
			} else if value.Valid {
				t.Requests = value.Int64
			}
		case token.FieldLastAccessAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field last_access_at", values[i])
			} else if value.Valid {
				t.LastAccessAt = value.Time
			}
		default:
			t.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Token.
// This includes values selected through modifiers, order, etc.
func (t *Token) Value(name string) (ent.Value, error) {
	return t.selectValues.Get(name)
}

// Update returns a builder for updating this Token.
// Note that you need to call Token.Unwrap() before calling this method if this Token
// was returned from a transaction, and the transaction was committed or rolled back.
func (t *Token) Update() *TokenUpdateOne {
	return NewTokenClient(t.config).UpdateOne(t)
}

// Unwrap unwraps the Token entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (t *Token) Unwrap() *Token {
	_tx, ok := t.config.driver.(*txDriver)
	if !ok {
		panic("ent: Token is not a transactional entity")
	}
	t.config.driver = _tx.drv
	return t
}

// String implements the fmt.Stringer.
func (t *Token) String() string {
	var builder strings.Builder
	builder.WriteString("Token(")
	builder.WriteString(fmt.Sprintf("id=%v, ", t.ID))
	builder.WriteString("created_at=")
	builder.WriteString(t.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(t.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("key_id=")
	builder.WriteString(fmt.Sprintf("%v", t.KeyID))
	builder.WriteString(", ")
	builder.WriteString("hash=")
	builder.WriteString(fmt.Sprintf("%v", t.Hash))
	builder.WriteString(", ")
	builder.WriteString("user=")
	builder.WriteString(t.User)
	builder.WriteString(", ")
	builder.WriteString("label=")
	builder.WriteString(t.Label)
	builder.WriteString(", ")
	builder.WriteString("path=")
	builder.WriteString(t.Path)
	builder.WriteString(", ")
	builder.WriteString("host=")
	builder.WriteString(t.Host)
	builder.WriteString(", ")
	builder.WriteString("headers=")
	builder.WriteString(fmt.Sprintf("%v", t.Headers))
	builder.WriteString(", ")
	builder.WriteString("requests=")
	builder.WriteString(fmt.Sprintf("%v", t.Requests))
	builder.WriteString(", ")
	builder.WriteString("last_access_at=")
	builder.WriteString(t.LastAccessAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Tokens is a parsable slice of Token.
type Tokens []*Token