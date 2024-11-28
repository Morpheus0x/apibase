package log

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	defaultLogLevel   = LevelNotice
	defaultTimeFormat = time.RFC3339
)

var (
	loggerMutex sync.RWMutex
	logger      = initDefaultLogger()
)

type SafeWriter struct {
	sync.Mutex
	Writer   io.Writer
	WithTime bool
}

// TODO: maybe change to global writer mutex to prevent edge case
// where mutliple targets may have different log order on high concurrency
type Logger struct {
	Level   Level
	TimeFmt string
	Writers []*SafeWriter
}

func initDefaultLogger() *Logger {
	return &Logger{
		Level:   defaultLogLevel,
		TimeFmt: defaultTimeFormat,
		Writers: []*SafeWriter{{
			Writer:   os.Stdout,
			WithTime: false,
		}},
	}
}

func ReplaceLogger(safeWriter ...*SafeWriter) {
	loggerMutex.Lock()
	//TODO: this
	loggerMutex.Unlock()
}

func AddLogger(safeWriter *SafeWriter) {
	loggerMutex.Lock()
	//TODO: this
	loggerMutex.Unlock()
}

func Log(l Level, s string) {
	loggerMutex.RLock()
	log := logger
	loggerMutex.RUnlock()
	log.printWithLevel(s, l)
}

func Logf(l Level, s string, a ...any) {
	loggerMutex.RLock()
	log := logger
	loggerMutex.RUnlock()
	log.printWithLevel(fmt.Sprintf(s, a...), l)
}

func LogMultiple(l Level, s []string) {
	loggerMutex.RLock()
	log := logger
	loggerMutex.RUnlock()
	log.printMultipleWithLevel(s, l)
}

func (log *Logger) printWithLevel(text string, l Level) {
	for _, w := range log.Writers {
		var logOutput []byte
		if w.WithTime {
			logOutput = []byte(fmt.Sprintf("%s %s %s\n", time.Now().Format(log.TimeFmt), l.String(), text))
		} else {
			logOutput = []byte(fmt.Sprintf("%s %s\n", l.String(), text))
		}
		w.Lock()
		w.Writer.Write(logOutput)
		w.Unlock()
	}
}

func (log *Logger) printMultipleWithLevel(text []string, l Level) {
	for _, w := range log.Writers {
		var logOutput []byte
		for _, s := range text {
			if w.WithTime {
				logOutput = append(logOutput, []byte(fmt.Sprintf("%s %s %s\n", time.Now().Format(log.TimeFmt), l.String(), s))...)
			} else {
				logOutput = append(logOutput, []byte(fmt.Sprintf("%s %s\n", l.String(), s))...)
			}
		}
		w.Lock()
		w.Writer.Write(logOutput)
		w.Unlock()
	}
}
