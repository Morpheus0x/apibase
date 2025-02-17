package web_response

const (
	QueryKeySuccess = "api_success"
	QueryKeyError   = "api_error"
)

// Standardized JSON API response
type JsonResponse[T any] struct {
	ResponseID ResponseId `json:"id"`
	Message    string     `json:"msg"`
	Data       T          `json:"data"`
}

type RedirectTarget struct {
	Referrer string `json:"ref"`
	Target   string `json:"target"`
}

// type HtmxResponse[T any] struct {
// 	HtmlTemplate string
// 	Data         T
// }
