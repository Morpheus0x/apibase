package helper

type ConfirmDefault bool

const (
	No  ConfirmDefault = false
	Yes ConfirmDefault = true
)

type ConfirmSensitivity uint

const (
	// Accepting any ordered subset of case insensitive letters of "yes" and "no", also enter for default
	Patient ConfirmSensitivity = iota
	// Requiring a fully typed out case insensitive "yes", "no" or "\n", otherwise ask repeatedly
	Picky
	// Requiring a fully typed out case insensitive "yes", otherwise immediately quit
	Strict
)
