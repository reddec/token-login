// Code generated by ogen, DO NOT EDIT.

package api

import (
	"time"
)

// Ref: #/components/schemas/Config
type Config struct {
	// Custom token description.
	Label OptString `json:"label"`
	// Allowed hosts. Supports globs. Empty means "allow all".
	Host OptString `json:"host"`
	// Allowed path. Supports globs. Empty means "allow all".
	Path OptString `json:"path"`
	// Custom headers which will be added after successfull authorization.
	Headers []NameValue `json:"headers"`
}

// GetLabel returns the value of Label.
func (s *Config) GetLabel() OptString {
	return s.Label
}

// GetHost returns the value of Host.
func (s *Config) GetHost() OptString {
	return s.Host
}

// GetPath returns the value of Path.
func (s *Config) GetPath() OptString {
	return s.Path
}

// GetHeaders returns the value of Headers.
func (s *Config) GetHeaders() []NameValue {
	return s.Headers
}

// SetLabel sets the value of Label.
func (s *Config) SetLabel(val OptString) {
	s.Label = val
}

// SetHost sets the value of Host.
func (s *Config) SetHost(val OptString) {
	s.Host = val
}

// SetPath sets the value of Path.
func (s *Config) SetPath(val OptString) {
	s.Path = val
}

// SetHeaders sets the value of Headers.
func (s *Config) SetHeaders(val []NameValue) {
	s.Headers = val
}

// Ref: #/components/schemas/Credential
type Credential struct {
	// Token ID.
	ID int `json:"id"`
	// Raw token key.
	Key string `json:"key"`
}

// GetID returns the value of ID.
func (s *Credential) GetID() int {
	return s.ID
}

// GetKey returns the value of Key.
func (s *Credential) GetKey() string {
	return s.Key
}

// SetID sets the value of ID.
func (s *Credential) SetID(val int) {
	s.ID = val
}

// SetKey sets the value of Key.
func (s *Credential) SetKey(val string) {
	s.Key = val
}

// DeleteTokenNoContent is response for DeleteToken operation.
type DeleteTokenNoContent struct{}

// Ref: #/components/schemas/NameValue
type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// GetName returns the value of Name.
func (s *NameValue) GetName() string {
	return s.Name
}

// GetValue returns the value of Value.
func (s *NameValue) GetValue() string {
	return s.Value
}

// SetName sets the value of Name.
func (s *NameValue) SetName(val string) {
	s.Name = val
}

// SetValue sets the value of Value.
func (s *NameValue) SetValue(val string) {
	s.Value = val
}

// NewOptDateTime returns new OptDateTime with value set to v.
func NewOptDateTime(v time.Time) OptDateTime {
	return OptDateTime{
		Value: v,
		Set:   true,
	}
}

// OptDateTime is optional time.Time.
type OptDateTime struct {
	Value time.Time
	Set   bool
}

// IsSet returns true if OptDateTime was set.
func (o OptDateTime) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptDateTime) Reset() {
	var v time.Time
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptDateTime) SetTo(v time.Time) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptDateTime) Get() (v time.Time, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptDateTime) Or(d time.Time) time.Time {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// NewOptString returns new OptString with value set to v.
func NewOptString(v string) OptString {
	return OptString{
		Value: v,
		Set:   true,
	}
}

// OptString is optional string.
type OptString struct {
	Value string
	Set   bool
}

// IsSet returns true if OptString was set.
func (o OptString) IsSet() bool { return o.Set }

// Reset unsets value.
func (o *OptString) Reset() {
	var v string
	o.Value = v
	o.Set = false
}

// SetTo sets value to v.
func (o *OptString) SetTo(v string) {
	o.Set = true
	o.Value = v
}

// Get returns value and boolean that denotes whether value was set.
func (o OptString) Get() (v string, ok bool) {
	if !o.Set {
		return v, false
	}
	return o.Value, true
}

// Or returns value if set, or given parameter if does not.
func (o OptString) Or(d string) string {
	if v, ok := o.Get(); ok {
		return v
	}
	return d
}

// Ref: #/components/schemas/Token
type Token struct {
	// Unique token ID.
	ID int `json:"id"`
	// Time when token was initially created.
	CreatedAt time.Time `json:"createdAt"`
	// Time when token was updated last time.
	UpdatedAt time.Time `json:"updatedAt"`
	// Tentative time when token was last time used.
	LastAccessAt OptDateTime `json:"lastAccessAt"`
	// Unique first several bytes for token which is used for fast identification.
	KeyID string `json:"keyID"`
	// User which created token.
	User string `json:"user"`
	// Custom token description.
	Label string `json:"label"`
	// Allowed hosts. Supports globs. Empty means "allow all".
	Host string `json:"host"`
	// Allowed path. Supports globs. Empty means "allow all".
	Path string `json:"path"`
	// Custom headers which will be added after successfull authorization.
	Headers []NameValue `json:"headers"`
	// Tentative number of requests used this token.
	Requests int64 `json:"requests"`
}

// GetID returns the value of ID.
func (s *Token) GetID() int {
	return s.ID
}

// GetCreatedAt returns the value of CreatedAt.
func (s *Token) GetCreatedAt() time.Time {
	return s.CreatedAt
}

// GetUpdatedAt returns the value of UpdatedAt.
func (s *Token) GetUpdatedAt() time.Time {
	return s.UpdatedAt
}

// GetLastAccessAt returns the value of LastAccessAt.
func (s *Token) GetLastAccessAt() OptDateTime {
	return s.LastAccessAt
}

// GetKeyID returns the value of KeyID.
func (s *Token) GetKeyID() string {
	return s.KeyID
}

// GetUser returns the value of User.
func (s *Token) GetUser() string {
	return s.User
}

// GetLabel returns the value of Label.
func (s *Token) GetLabel() string {
	return s.Label
}

// GetHost returns the value of Host.
func (s *Token) GetHost() string {
	return s.Host
}

// GetPath returns the value of Path.
func (s *Token) GetPath() string {
	return s.Path
}

// GetHeaders returns the value of Headers.
func (s *Token) GetHeaders() []NameValue {
	return s.Headers
}

// GetRequests returns the value of Requests.
func (s *Token) GetRequests() int64 {
	return s.Requests
}

// SetID sets the value of ID.
func (s *Token) SetID(val int) {
	s.ID = val
}

// SetCreatedAt sets the value of CreatedAt.
func (s *Token) SetCreatedAt(val time.Time) {
	s.CreatedAt = val
}

// SetUpdatedAt sets the value of UpdatedAt.
func (s *Token) SetUpdatedAt(val time.Time) {
	s.UpdatedAt = val
}

// SetLastAccessAt sets the value of LastAccessAt.
func (s *Token) SetLastAccessAt(val OptDateTime) {
	s.LastAccessAt = val
}

// SetKeyID sets the value of KeyID.
func (s *Token) SetKeyID(val string) {
	s.KeyID = val
}

// SetUser sets the value of User.
func (s *Token) SetUser(val string) {
	s.User = val
}

// SetLabel sets the value of Label.
func (s *Token) SetLabel(val string) {
	s.Label = val
}

// SetHost sets the value of Host.
func (s *Token) SetHost(val string) {
	s.Host = val
}

// SetPath sets the value of Path.
func (s *Token) SetPath(val string) {
	s.Path = val
}

// SetHeaders sets the value of Headers.
func (s *Token) SetHeaders(val []NameValue) {
	s.Headers = val
}

// SetRequests sets the value of Requests.
func (s *Token) SetRequests(val int64) {
	s.Requests = val
}

// UpdateTokenNoContent is response for UpdateToken operation.
type UpdateTokenNoContent struct{}
