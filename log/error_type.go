package log

type errorType struct {
	Name string
}

var DefaultError = RegisterErrType("ErrDefault")

func RegisterErrType(name string) *errorType {
	return &errorType{Name: name}
}

func (et *errorType) String() string {
	return et.Name
}
