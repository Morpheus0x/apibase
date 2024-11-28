package log

import (
	"fmt"
	"runtime"
)

// Source: https://stackoverflow.com/a/46289376
func traceError() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
}
