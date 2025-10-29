package errx

import (
	"fmt"
)

// Custom error type, implementing built-in error interface

// Sources:
// https://www.digitalocean.com/community/tutorials/creating-custom-errors-in-go
// https://freedium.cfd/https://mdcfrancis.medium.com/tracing-errors-in-go-using-custom-error-types-9aaf3bba1a64

type BaseErrorType struct {
	text string
}

func (e *BaseErrorType) Error() string {
	return e.text
}

type BaseError struct {
	errType error
	text    string
	nested  error
}

// Get BaseError text and append nested error(s), if any
func (e *BaseError) Error() string {
	out := ""
	if e.errType != nil {
		out += e.errType.Error()
	}
	if e.text != "" {
		if out != "" {
			out += ": "
		}
		out += e.text
	}
	if e.nested != nil {
		if out != "" {
			out += ": "
		}
		// if nested error also is BaseError this prints recursively
		out += e.nested.Error()
	}
	return out
}

// get nested error
func (e *BaseError) Unwrap() error {
	return e.nested
}

// Compare if the error is of same type as target,
// if both have no type, compare the error message itself
func (e *BaseError) Is(target error) bool {
	errType, ok := target.(*BaseErrorType)
	if ok {
		return e.errType == errType
	}
	isErr, ok := target.(*BaseError)
	if !ok {
		return false
	}
	if e.errType == nil && isErr.errType == nil {
		return e == target
	}
	return e.errType == isErr.errType
}

// func (e *BaseError) Type() error {
// 	return e.errType
// }

func NewType(descr string) error {
	return &BaseErrorType{
		text: descr,
	}
}

func New(text string) error {
	return &BaseError{
		text: text,
	}
}

func Newf(text string, a ...any) error {
	return &BaseError{
		text: fmt.Sprintf(text, a...),
	}
}

func NewWithType(errType error, text string) error {
	return &BaseError{
		errType: errType,
		text:    text,
	}
}

func NewWithTypef(errType error, text string, a ...any) error {
	return &BaseError{
		errType: errType,
		text:    fmt.Sprintf(text, a...),
	}
}

// create new error which wraps err
func Wrap(err error, text string) error {
	return &BaseError{
		text:   text,
		nested: err,
	}
}

// create new error which wraps err, with formatting
func Wrapf(err error, text string, a ...any) error {
	return &BaseError{
		text:   fmt.Sprintf(text, a...),
		nested: err,
	}
}

// create new error which wraps err
func WrapWithType(errType error, err error, text string) error {
	return &BaseError{
		errType: errType,
		text:    text,
		nested:  err,
	}
}

// create new error which wraps err
func WrapWithTypef(errType error, err error, text string, a ...any) error {
	return &BaseError{
		errType: errType,
		text:    fmt.Sprintf(text, a...),
		nested:  err,
	}
}
