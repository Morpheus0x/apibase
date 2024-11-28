package log

import (
	"fmt"
)

func ErrorNil() *Error {
	return &Error{isNil: true}
}

func NewError(text string) *Error {
	return &Error{
		errType:  DefaultError,
		severity: LevelError,
		content: []ErrorContent{{
			text:  text,
			trace: traceError(),
		}},
		isNil:     false,
		wasLogged: false,
	}
}

// Create new Error
func NewErrorf(format string, a ...any) *Error {
	text := ""
	if format != "" {
		text = fmt.Sprintf(format, a...)
	}
	return NewError(text)
}

func NewErrorWithType(t *errorType, text string) *Error {
	return &Error{
		errType:  t,
		severity: LevelError,
		content: []ErrorContent{{
			text:  text,
			trace: traceError(),
		}},
		isNil:     false,
		wasLogged: false,
	}
}

func NewErrorWithTypef(t *errorType, format string, a ...any) *Error {
	text := ""
	if format != "" {
		text = fmt.Sprintf(format, a...)
	}
	return NewErrorWithType(t, text)
}
