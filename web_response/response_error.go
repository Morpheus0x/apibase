package web_response

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Error ID for localized errors on frontend
//
//go:generate stringer -type ResponseErrorId -output ./stringer_ResponseErrorId.go
type ResponseErrorId uint

const (
	RespErrNone ResponseErrorId = iota
	RespErrUndefined
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
	// Only append here to not break existing frontend error IDs
)

// func (respErrId ResponseErrorId) Int() int

type ResponseError struct {
	httpStatus int
	errorId    ResponseErrorId
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

func (e ResponseError) GetErrorId() ResponseErrorId {
	return e.errorId
}

func (e ResponseError) SendJson(c echo.Context) error {
	httpStatus := http.StatusInternalServerError
	if e.httpStatus != 0 {
		httpStatus = e.httpStatus
	}
	return c.JSON(httpStatus, JsonResponse[struct{}]{ErrorID: e.errorId})
}

func (e ResponseError) SendJsonWithStatus(c echo.Context, httpStatus int) error {
	return c.JSON(httpStatus, JsonResponse[struct{}]{ErrorID: e.errorId})
}

func NewError(errorId ResponseErrorId, err error) error {
	return &ResponseError{
		httpStatus: 0,
		errorId:    errorId,
		nested:     err,
	}
}

func NewErrorWithStatus(status int, errorId ResponseErrorId, err error) error {
	return &ResponseError{
		httpStatus: status,
		errorId:    errorId,
		nested:     err,
	}
}
