package log

//go:generate stringer -type ErrorType -output ./stringer_errortype.go
type ErrorType uint

const (
	// Default Error, should not be used
	ErrUndefined ErrorType = iota

	// General Errors
	// Provided string is empty
	ErrEmptyString
	// Provided array is empty
	ErrEmptyArray

	// Package Errors: web
	// Unknown HTTP Method
	ErrWebUnknownMethod
	// ApiServer struct is not initialized
	ErrWebApiNotInit
	// Error initializing ApiServer
	ErrWebApiInit
	// ApiServer group already exists
	ErrWebGroupExists
	// ApiServer group exists but is required
	ErrWebGroupNotExists
	// Unable to start ApiServer with specified bind
	ErrWebBind
	// Token could not be parsed from cookie
	ErrTokenValidate
)

type Err struct {
	Type  ErrorType
	text  string
	isNil bool
}

func (e Err) Text() string {
	return e.text
}

func (e Err) IsNil() bool {
	return e.isNil
}

func (e Err) Is(errType ErrorType) bool {
	return e.Type == errType
}
