package web_response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/log"
)

// Error ID for localized errors on frontend
//
//go:generate stringer -type ResponseId -output ./stringer_ResponseId.go
type ResponseId uint

const (
	RespSccsGeneric ResponseId = iota
	RespSccsLogin
	RespSccsLogout
	RespSccsSignup
	RespScssSignupEmailConfirm
	RespSccsAlreadyLoggedIn
	RespErrUndefined
	RespErrUnknownInternal
	RespErrCsrfInvalid
	RespErrUserDoesNotExist
	RespErrUserNoRoles
	RespErrMissingInput
	RespErrJwtAccessTokenSigning
	RespErrJwtAccessTokenParsing
	RespErrJwtRefreshTokenCreate
	RespErrJwtRefreshTokenSigning
	RespErrJwtRefreshTokenUpdate
	RespErrJwtRefreshTokenParsing
	RespErrJwtRefreshTokenClaims
	RespErrJwtRefreshTokenInvalid
	RespErrJwtRefreshTokenExpired
	RespErrJwtRefreshTokenVerifyErr
	RespErrJwtRefreshTokenVerifyInvalid
	RespErrOauthCallbackCompleteAuth
	RespErrOauthCallbackUnknownError
	RespErrAuthLoginUnknownError
	RespErrAuthLoginNotLocal
	RespErrAuthSignupUnknownError
	RespErrAuthLogoutUnknownError
	RespErrLoginNoUser
	RespErrLoginComparePassword
	RespErrLoginWrongPassword
	RespErrSignupPasswordMismatch
	RespErrSignupPasswordHash
	RespErrSignupUserExists
	RespErrSignupNewUserOrg
	RespErrSignupUserCreate
	RespErrHookPreLogin
	RespErrHookPostLogin
	RespErrHookPreSignup
	RespErrHookSignupDefaultRole
	RespErrOauthReferrerParsing
	RespErrOauthMarshalState
	RespErrGetAccessClaims
	// Only append here to not break existing frontend error IDs
)

func SendJsonErrorResponse(c echo.Context, httpStatus int, responseId ResponseId) error {
	if httpStatus < 400 && strings.HasPrefix(responseId.String(), "RespErr") {
		log.Logf(log.LevelWarning, "web_response.SendJsonErrorResponse called with error ResponseId (%s) but with non-error http status code (%d), assuming StatusInternalServerError", responseId.String(), httpStatus)
		httpStatus = http.StatusInternalServerError
	} else if httpStatus >= 400 && !strings.HasPrefix(responseId.String(), "RespErr") {
		log.Logf(log.LevelWarning, "web_response.SendJsonErrorResponse called with error http status code (%d) but with non-error ResponseId (%s), assuming RespErrUndefined", httpStatus, responseId.String())
		responseId = RespErrUndefined
	}
	if httpStatus >= 300 && httpStatus <= 399 {
		log.Logf(log.LevelWarning, "invalid http status code (%d) for JSON response, assuming StatusOK", httpStatus)
		httpStatus = http.StatusOK
	}
	return c.JSON(httpStatus, JsonResponse[struct{}]{ResponseID: responseId})
}

type ResponseError struct {
	httpStatus int
	errorId    ResponseId
	nested     error
}

func (e ResponseError) Error() string {
	out := e.errorId.String()
	if e.httpStatus != 0 {
		out += fmt.Sprintf("(HTTP Status %d)", e.httpStatus)
	}
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

func (e ResponseError) GetErrorId() ResponseId {
	return e.errorId
}

func (e ResponseError) GetResponse() (JsonResponse[struct{}], int) {
	httpStatus := http.StatusInternalServerError
	if e.httpStatus != 0 {
		httpStatus = e.httpStatus
	}
	return JsonResponse[struct{}]{
		ResponseID: e.errorId,
		Message:    e.Error(),
	}, httpStatus
}

func (e ResponseError) SendJson(c echo.Context) error {
	httpStatus := http.StatusInternalServerError
	if e.httpStatus != 0 {
		httpStatus = e.httpStatus
	}
	return c.JSON(httpStatus, JsonResponse[struct{}]{ResponseID: e.errorId})
}

func (e ResponseError) SendJsonWithStatus(c echo.Context, httpStatus int) error {
	return c.JSON(httpStatus, JsonResponse[struct{}]{ResponseID: e.errorId})
}

func NewError(errorId ResponseId, err error) error {
	if !strings.HasPrefix(errorId.String(), "RespErr") {
		log.Logf(log.LevelWarning, "web_response.NewError called with non-error ResponseId (%s), assuming RespErrUndefined", errorId.String())
		errorId = RespErrUndefined
	}
	return &ResponseError{
		httpStatus: 0,
		errorId:    errorId,
		nested:     err,
	}
}

func NewErrorWithStatus(status int, errorId ResponseId, err error) error {
	if !strings.HasPrefix(errorId.String(), "RespErr") {
		log.Logf(log.LevelWarning, "web_response.NewErrorWithStatus called with non-error ResponseId (%s), assuming RespErrUndefined", errorId.String())
		errorId = RespErrUndefined
	}
	return &ResponseError{
		httpStatus: status,
		errorId:    errorId,
		nested:     err,
	}
}
