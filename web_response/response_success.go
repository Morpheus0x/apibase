package web_response

// Success ID for localized message on frontend
//
//go:generate stringer -type ResponseSuccessId -output ./stringer_ResponseSuccessId.go
type ResponseSuccessId uint

const (
	RespSccsGeneric ResponseSuccessId = iota
	RespSccsLogin
	RespSccsLogout
	// Only append here to not break existing frontend success IDs
)
