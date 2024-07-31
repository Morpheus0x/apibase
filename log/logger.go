package log

//go:generate stringer -type Level -output ./stringer_level.go
type Level uint

const (
	Debug Level = iota
	Info
	Notice
	Warning
	Error
)
