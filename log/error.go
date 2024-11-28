package log

import (
	"fmt"
	"strings"
)

type ErrorContent struct {
	text  string
	trace string
}

type Error struct {
	errType   *errorType
	severity  Level
	content   []ErrorContent
	isNil     bool
	wasLogged bool
}

// Error Extending Functions

func (e *Error) Severity(level Level) *Error {
	e.severity = level
	return e
}

func (e *Error) Extend(text string) *Error {
	e.content = append(e.content, ErrorContent{
		text:  text,
		trace: traceError(),
	})
	e.wasLogged = false
	return e
}

func (e *Error) Log() *Error {
	if e.wasLogged {
		return e
	}
	e.wasLogged = true
	if e.errType == DefaultError {
		Log(e.severity, e.String())
	} else {
		Log(e.severity, fmt.Sprintf("%s: %s", e.errType.String(), e.String()))
	}
	return e
}

func (e *Error) LogWithTrace() *Error {
	if e.wasLogged {
		return e
	}
	e.wasLogged = true
	if e.errType == DefaultError {

	} else {

	}
	LogMultiple(e.severity, e.StringWithTrace())
	return e
}

// Error Output Functions

func (e *Error) String() string {
	var sb strings.Builder
	if e.errType != DefaultError {
		sb.WriteString(e.errType.String())
	}
	for i := len(e.content) - 1; i >= 0; i-- {
		if e.content[i].text == "" {
			continue
		}
		if sb.Len() != 0 {
			sb.WriteString(": ")
		}
		sb.WriteString(e.content[i].text)
	}
	return sb.String()
}

func (e *Error) StringWithTrace() []string {
	var sb strings.Builder
	if e.errType != DefaultError {
		sb.WriteString(e.errType.String())
	}
	var trace []string
	contentLen := len(e.content) - 1
	for cnt, content := range e.content {
		trace = append(trace, content.trace)
		if e.content[contentLen-cnt].text == "" {
			continue
		}
		if sb.Len() != 0 {
			sb.WriteString(": ")
		}
		sb.WriteString(e.content[contentLen-cnt].text)
	}
	return append([]string{sb.String()}, trace...)
}

func (e *Error) IsNil() bool {
	return e.isNil
}

func (e *Error) IsType(errType *errorType) bool {
	return e.errType == errType
}

// Log error to default logger and panic
func (e *Error) Panic() {
	e.severity = LevelCritical
	e.Log()
	panic(e.String())
}
