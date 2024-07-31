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
