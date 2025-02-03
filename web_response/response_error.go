package web_response

// Error ID for localized errors on frontend
//
//go:generate stringer -type ResponseErrorId -output ./stringer_ResponseErrorId.go
type ResponseErrorId uint

const (
	RespErrNone ResponseErrorId = iota
	RespErrUndefined
	// Only append here to not break existing frontend error IDs
)

// func (respErrId ResponseErrorId) Int() int

type ResponseError struct {
	errorId ResponseErrorId
	nested  error
}

func (e ResponseError) Error() string {
	out := e.errorId.String()
	if e.nested != nil {
		out += ": " + e.nested.Error()
	}
	return out
}

// get nested error
func (e ResponseError) Unwrap() error {
	return e.nested
}

// Compare if the error is of type ResponseError and has the same errorId
func (e ResponseError) Is(target error) bool {
	isErr, ok := target.(*ResponseError)
	if !ok {
		return false
	}
	return e.errorId == isErr.errorId
}

func (e ResponseError) GetErrorId() ResponseErrorId {
	return e.errorId
}

func NewResponseError(errorId ResponseErrorId, err error) error {
	return &ResponseError{
		errorId: errorId,
		nested:  err,
	}
}
