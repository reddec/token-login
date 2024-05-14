// Code generated by ogen, DO NOT EDIT.

package api

import (
	"fmt"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/validate"
)

func (s *Config) Validate() error {
	if s == nil {
		return validate.ErrNilPointer
	}

	var failures []validate.FieldError
	if err := func() error {
		if value, ok := s.Label.Get(); ok {
			if err := func() error {
				if err := (validate.String{
					MinLength:    0,
					MinLengthSet: false,
					MaxLength:    127,
					MaxLengthSet: true,
					Email:        false,
					Hostname:     false,
					Regex:        nil,
				}).Validate(string(value)); err != nil {
					return errors.Wrap(err, "string")
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		failures = append(failures, validate.FieldError{
			Name:  "label",
			Error: err,
		})
	}
	if err := func() error {
		if value, ok := s.Host.Get(); ok {
			if err := func() error {
				if err := (validate.String{
					MinLength:    0,
					MinLengthSet: false,
					MaxLength:    255,
					MaxLengthSet: true,
					Email:        false,
					Hostname:     false,
					Regex:        nil,
				}).Validate(string(value)); err != nil {
					return errors.Wrap(err, "string")
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		failures = append(failures, validate.FieldError{
			Name:  "host",
			Error: err,
		})
	}
	if err := func() error {
		if value, ok := s.Path.Get(); ok {
			if err := func() error {
				if err := (validate.String{
					MinLength:    0,
					MinLengthSet: false,
					MaxLength:    2048,
					MaxLengthSet: true,
					Email:        false,
					Hostname:     false,
					Regex:        nil,
				}).Validate(string(value)); err != nil {
					return errors.Wrap(err, "string")
				}
				return nil
			}(); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		failures = append(failures, validate.FieldError{
			Name:  "path",
			Error: err,
		})
	}
	if err := func() error {
		if err := (validate.Array{
			MinLength:    0,
			MinLengthSet: false,
			MaxLength:    20,
			MaxLengthSet: true,
		}).ValidateLength(len(s.Headers)); err != nil {
			return errors.Wrap(err, "array")
		}
		var failures []validate.FieldError
		for i, elem := range s.Headers {
			if err := func() error {
				if err := elem.Validate(); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				failures = append(failures, validate.FieldError{
					Name:  fmt.Sprintf("[%d]", i),
					Error: err,
				})
			}
		}
		if len(failures) > 0 {
			return &validate.Error{Fields: failures}
		}
		return nil
	}(); err != nil {
		failures = append(failures, validate.FieldError{
			Name:  "headers",
			Error: err,
		})
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}
	return nil
}

func (s *NameValue) Validate() error {
	if s == nil {
		return validate.ErrNilPointer
	}

	var failures []validate.FieldError
	if err := func() error {
		if err := (validate.String{
			MinLength:    0,
			MinLengthSet: false,
			MaxLength:    50,
			MaxLengthSet: true,
			Email:        false,
			Hostname:     false,
			Regex:        nil,
		}).Validate(string(s.Name)); err != nil {
			return errors.Wrap(err, "string")
		}
		return nil
	}(); err != nil {
		failures = append(failures, validate.FieldError{
			Name:  "name",
			Error: err,
		})
	}
	if err := func() error {
		if err := (validate.String{
			MinLength:    0,
			MinLengthSet: false,
			MaxLength:    1024,
			MaxLengthSet: true,
			Email:        false,
			Hostname:     false,
			Regex:        nil,
		}).Validate(string(s.Value)); err != nil {
			return errors.Wrap(err, "string")
		}
		return nil
	}(); err != nil {
		failures = append(failures, validate.FieldError{
			Name:  "value",
			Error: err,
		})
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}
	return nil
}

func (s *Token) Validate() error {
	if s == nil {
		return validate.ErrNilPointer
	}

	var failures []validate.FieldError
	if err := func() error {
		var failures []validate.FieldError
		for i, elem := range s.Headers {
			if err := func() error {
				if err := elem.Validate(); err != nil {
					return err
				}
				return nil
			}(); err != nil {
				failures = append(failures, validate.FieldError{
					Name:  fmt.Sprintf("[%d]", i),
					Error: err,
				})
			}
		}
		if len(failures) > 0 {
			return &validate.Error{Fields: failures}
		}
		return nil
	}(); err != nil {
		failures = append(failures, validate.FieldError{
			Name:  "headers",
			Error: err,
		})
	}
	if len(failures) > 0 {
		return &validate.Error{Fields: failures}
	}
	return nil
}
