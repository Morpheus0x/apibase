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
		Writers: DefaultSafeWriter(false),
	}
}

// write to stdout, specify if time should also be printed
func DefaultSafeWriter(withTime bool) []*SafeWriter {
	return []*SafeWriter{{
		Writer:   os.Stdout,
		WithTime: withTime,
	}}
}

// specify desired log verbosity, desired time format e.g. time.RFC3339 (always specify this, even if SafeWriter is set to WithTime = false)
// and the desired log output destination (use DefaultSafeWriter() for stdout)
func ReplaceLogger(level Level, timeFmt string, safeWriters []*SafeWriter) {
	loggerMutex.Lock()
	logger = &Logger{
		Level:   level,
		TimeFmt: timeFmt,
		Writers: safeWriters,
	}
	loggerMutex.Unlock()
	// TODO: tryout this new logger to make sure it doesn't panic, log to info
	// logger.Writers[len(logger.Writers)-1]
}

// add an additional log output destination
func AddLogger(safeWriter *SafeWriter) {
	loggerMutex.Lock()
	logger.Writers = append(logger.Writers, safeWriter)
	loggerMutex.Unlock()
	// TODO: tryout this new logger to make sure it doesn't panic, log to info
}

// set log level, this should be done only once in the beginning
func SetLogLevel(level Level) {
	loggerMutex.Lock()
	logger.Level = level
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

func (log *Logger) printWithLevel(text string, level Level) {
	if level < log.Level {
		// fmt.Printf("level(%d) < log.Level(%d) = %v\n", level, log.Level, level < log.Level)
		// log message is too verbose level is higher than set in logger, ignoring
		return
	}
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

func (log *Logger) printMultipleWithLevel(text []string, level Level) {
	if level < log.Level {
		// fmt.Printf("level(%d) < log.Level(%d) = %v\n", level, log.Level, level < log.Level)
		// log message is too verbose level is higher than set in logger, ignoring
		return
	}
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
