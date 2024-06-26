// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/reddec/token-login/internal/ent/token"
	"github.com/reddec/token-login/internal/types"
)

// TokenCreate is the builder for creating a Token entity.
type TokenCreate struct {
	config
	mutation *TokenMutation
	hooks    []Hook
}

// SetCreatedAt sets the "created_at" field.
func (tc *TokenCreate) SetCreatedAt(t time.Time) *TokenCreate {
	tc.mutation.SetCreatedAt(t)
	return tc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (tc *TokenCreate) SetNillableCreatedAt(t *time.Time) *TokenCreate {
	if t != nil {
		tc.SetCreatedAt(*t)
	}
	return tc
}

// SetUpdatedAt sets the "updated_at" field.
func (tc *TokenCreate) SetUpdatedAt(t time.Time) *TokenCreate {
	tc.mutation.SetUpdatedAt(t)
	return tc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (tc *TokenCreate) SetNillableUpdatedAt(t *time.Time) *TokenCreate {
	if t != nil {
		tc.SetUpdatedAt(*t)
	}
	return tc
}

// SetKeyID sets the "key_id" field.
func (tc *TokenCreate) SetKeyID(ti *types.KeyID) *TokenCreate {
	tc.mutation.SetKeyID(ti)
	return tc
}

// SetHash sets the "hash" field.
func (tc *TokenCreate) SetHash(b []byte) *TokenCreate {
	tc.mutation.SetHash(b)
	return tc
}

// SetUser sets the "user" field.
func (tc *TokenCreate) SetUser(s string) *TokenCreate {
	tc.mutation.SetUser(s)
	return tc
}

// SetLabel sets the "label" field.
func (tc *TokenCreate) SetLabel(s string) *TokenCreate {
	tc.mutation.SetLabel(s)
	return tc
}

// SetNillableLabel sets the "label" field if the given value is not nil.
func (tc *TokenCreate) SetNillableLabel(s *string) *TokenCreate {
	if s != nil {
		tc.SetLabel(*s)
	}
	return tc
}

// SetPath sets the "path" field.
func (tc *TokenCreate) SetPath(s string) *TokenCreate {
	tc.mutation.SetPath(s)
	return tc
}

// SetNillablePath sets the "path" field if the given value is not nil.
func (tc *TokenCreate) SetNillablePath(s *string) *TokenCreate {
	if s != nil {
		tc.SetPath(*s)
	}
	return tc
}

// SetHost sets the "host" field.
func (tc *TokenCreate) SetHost(s string) *TokenCreate {
	tc.mutation.SetHost(s)
	return tc
}

// SetNillableHost sets the "host" field if the given value is not nil.
func (tc *TokenCreate) SetNillableHost(s *string) *TokenCreate {
	if s != nil {
		tc.SetHost(*s)
	}
	return tc
}

// SetHeaders sets the "headers" field.
func (tc *TokenCreate) SetHeaders(t types.Headers) *TokenCreate {
	tc.mutation.SetHeaders(t)
	return tc
}

// SetRequests sets the "requests" field.
func (tc *TokenCreate) SetRequests(i int64) *TokenCreate {
	tc.mutation.SetRequests(i)
	return tc
}

// SetNillableRequests sets the "requests" field if the given value is not nil.
func (tc *TokenCreate) SetNillableRequests(i *int64) *TokenCreate {
	if i != nil {
		tc.SetRequests(*i)
	}
	return tc
}

// SetLastAccessAt sets the "last_access_at" field.
func (tc *TokenCreate) SetLastAccessAt(t time.Time) *TokenCreate {
	tc.mutation.SetLastAccessAt(t)
	return tc
}

// SetNillableLastAccessAt sets the "last_access_at" field if the given value is not nil.
func (tc *TokenCreate) SetNillableLastAccessAt(t *time.Time) *TokenCreate {
	if t != nil {
		tc.SetLastAccessAt(*t)
	}
	return tc
}

// SetID sets the "id" field.
func (tc *TokenCreate) SetID(i int) *TokenCreate {
	tc.mutation.SetID(i)
	return tc
}

// Mutation returns the TokenMutation object of the builder.
func (tc *TokenCreate) Mutation() *TokenMutation {
	return tc.mutation
}

// Save creates the Token in the database.
func (tc *TokenCreate) Save(ctx context.Context) (*Token, error) {
	tc.defaults()
	return withHooks(ctx, tc.sqlSave, tc.mutation, tc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (tc *TokenCreate) SaveX(ctx context.Context) *Token {
	v, err := tc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tc *TokenCreate) Exec(ctx context.Context) error {
	_, err := tc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tc *TokenCreate) ExecX(ctx context.Context) {
	if err := tc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tc *TokenCreate) defaults() {
	if _, ok := tc.mutation.CreatedAt(); !ok {
		v := token.DefaultCreatedAt()
		tc.mutation.SetCreatedAt(v)
	}
	if _, ok := tc.mutation.UpdatedAt(); !ok {
		v := token.DefaultUpdatedAt()
		tc.mutation.SetUpdatedAt(v)
	}
	if _, ok := tc.mutation.Label(); !ok {
		v := token.DefaultLabel
		tc.mutation.SetLabel(v)
	}
	if _, ok := tc.mutation.Path(); !ok {
		v := token.DefaultPath
		tc.mutation.SetPath(v)
	}
	if _, ok := tc.mutation.Host(); !ok {
		v := token.DefaultHost
		tc.mutation.SetHost(v)
	}
	if _, ok := tc.mutation.Requests(); !ok {
		v := token.DefaultRequests
		tc.mutation.SetRequests(v)
	}
	if _, ok := tc.mutation.LastAccessAt(); !ok {
		v := token.DefaultLastAccessAt()
		tc.mutation.SetLastAccessAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tc *TokenCreate) check() error {
	if _, ok := tc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Token.created_at"`)}
	}
	if _, ok := tc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Token.updated_at"`)}
	}
	if _, ok := tc.mutation.KeyID(); !ok {
		return &ValidationError{Name: "key_id", err: errors.New(`ent: missing required field "Token.key_id"`)}
	}
	if _, ok := tc.mutation.Hash(); !ok {
		return &ValidationError{Name: "hash", err: errors.New(`ent: missing required field "Token.hash"`)}
	}
	if v, ok := tc.mutation.Hash(); ok {
		if err := token.HashValidator(v); err != nil {
			return &ValidationError{Name: "hash", err: fmt.Errorf(`ent: validator failed for field "Token.hash": %w`, err)}
		}
	}
	if _, ok := tc.mutation.User(); !ok {
		return &ValidationError{Name: "user", err: errors.New(`ent: missing required field "Token.user"`)}
	}
	if _, ok := tc.mutation.Label(); !ok {
		return &ValidationError{Name: "label", err: errors.New(`ent: missing required field "Token.label"`)}
	}
	if _, ok := tc.mutation.Path(); !ok {
		return &ValidationError{Name: "path", err: errors.New(`ent: missing required field "Token.path"`)}
	}
	if _, ok := tc.mutation.Host(); !ok {
		return &ValidationError{Name: "host", err: errors.New(`ent: missing required field "Token.host"`)}
	}
	if _, ok := tc.mutation.Requests(); !ok {
		return &ValidationError{Name: "requests", err: errors.New(`ent: missing required field "Token.requests"`)}
	}
	if _, ok := tc.mutation.LastAccessAt(); !ok {
		return &ValidationError{Name: "last_access_at", err: errors.New(`ent: missing required field "Token.last_access_at"`)}
	}
	return nil
}

func (tc *TokenCreate) sqlSave(ctx context.Context) (*Token, error) {
	if err := tc.check(); err != nil {
		return nil, err
	}
	_node, _spec, err := tc.createSpec()
	if err != nil {
		return nil, err
	}
	if err := sqlgraph.CreateNode(ctx, tc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != _node.ID {
		id := _spec.ID.Value.(int64)
		_node.ID = int(id)
	}
	tc.mutation.id = &_node.ID
	tc.mutation.done = true
	return _node, nil
}

func (tc *TokenCreate) createSpec() (*Token, *sqlgraph.CreateSpec, error) {
	var (
		_node = &Token{config: tc.config}
		_spec = sqlgraph.NewCreateSpec(token.Table, sqlgraph.NewFieldSpec(token.FieldID, field.TypeInt))
	)
	if id, ok := tc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = id
	}
	if value, ok := tc.mutation.CreatedAt(); ok {
		_spec.SetField(token.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := tc.mutation.UpdatedAt(); ok {
		_spec.SetField(token.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := tc.mutation.KeyID(); ok {
		vv, err := token.ValueScanner.KeyID.Value(value)
		if err != nil {
			return nil, nil, err
		}
		_spec.SetField(token.FieldKeyID, field.TypeString, vv)
		_node.KeyID = value
	}
	if value, ok := tc.mutation.Hash(); ok {
		_spec.SetField(token.FieldHash, field.TypeBytes, value)
		_node.Hash = value
	}
	if value, ok := tc.mutation.User(); ok {
		_spec.SetField(token.FieldUser, field.TypeString, value)
		_node.User = value
	}
	if value, ok := tc.mutation.Label(); ok {
		_spec.SetField(token.FieldLabel, field.TypeString, value)
		_node.Label = value
	}
	if value, ok := tc.mutation.Path(); ok {
		_spec.SetField(token.FieldPath, field.TypeString, value)
		_node.Path = value
	}
	if value, ok := tc.mutation.Host(); ok {
		_spec.SetField(token.FieldHost, field.TypeString, value)
		_node.Host = value
	}
	if value, ok := tc.mutation.Headers(); ok {
		_spec.SetField(token.FieldHeaders, field.TypeJSON, value)
		_node.Headers = value
	}
	if value, ok := tc.mutation.Requests(); ok {
		_spec.SetField(token.FieldRequests, field.TypeInt64, value)
		_node.Requests = value
	}
	if value, ok := tc.mutation.LastAccessAt(); ok {
		_spec.SetField(token.FieldLastAccessAt, field.TypeTime, value)
		_node.LastAccessAt = value
	}
	return _node, _spec, nil
}

// TokenCreateBulk is the builder for creating many Token entities in bulk.
type TokenCreateBulk struct {
	config
	err      error
	builders []*TokenCreate
}

// Save creates the Token entities in the database.
func (tcb *TokenCreateBulk) Save(ctx context.Context) ([]*Token, error) {
	if tcb.err != nil {
		return nil, tcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(tcb.builders))
	nodes := make([]*Token, len(tcb.builders))
	mutators := make([]Mutator, len(tcb.builders))
	for i := range tcb.builders {
		func(i int, root context.Context) {
			builder := tcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TokenMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i], err = builder.createSpec()
				if err != nil {
					return nil, err
				}
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, tcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				if specs[i].ID.Value != nil && nodes[i].ID == 0 {
					id := specs[i].ID.Value.(int64)
					nodes[i].ID = int(id)
				}
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, tcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tcb *TokenCreateBulk) SaveX(ctx context.Context) []*Token {
	v, err := tcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tcb *TokenCreateBulk) Exec(ctx context.Context) error {
	_, err := tcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcb *TokenCreateBulk) ExecX(ctx context.Context) {
	if err := tcb.Exec(ctx); err != nil {
		panic(err)
	}
}
