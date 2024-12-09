package log

import (
	"fmt"
)

type genericLogger struct {
	level Level
}

func (l *genericLogger) Printf(format string, v ...any) {
	Log(l.level, fmt.Sprintf(format, v...))
}

func NewGenericLogger(l Level) *genericLogger {
	return &genericLogger{level: l}
}
