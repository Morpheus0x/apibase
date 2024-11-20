package log

import "fmt"

func Panic(t ErrorType, format string, a ...any) {
	// TODO: use global variable initialized by Init() or custom func to select where to print to (stdout, file, ...)
	fmt.Printf("Fatal Error (%s) occured: %s", t.String(), fmt.Sprintf(format, a...))
	panic(t)
}

func ErrorNew(t ErrorType, format string, a ...any) Err {
	text := ""
	if format != "" {
		text = fmt.Sprintf(format, a...)
	}
	return Err{Type: t, text: text, isNil: false}
}

func ErrorNil() Err {
	return Err{isNil: true}
}
