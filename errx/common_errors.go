package errx

var (
	ErrNotImplemented    = NewType("Not Implemented - WIP")
	ErrSomeMinorOccurred = NewType("Some minor error(s) occerred: ")
)
