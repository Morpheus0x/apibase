package web_response

const (
	QueryKeySuccess = "api_success"
	QueryKeyError   = "api_error"
)

// Standardized JSON API response
type JsonResponse[T any] struct {
	ErrorID ResponseErrorId `json:"err"`
	Message string          `json:"msg"`
	Data    T               `json:"data"`
}

// type HtmxResponse[T any] struct {
// 	HtmlTemplate string
// 	Data         T
// }
