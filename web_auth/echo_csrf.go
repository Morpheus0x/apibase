package web_auth

import (
	"reflect"

	"github.com/labstack/echo/v4"
	t "gopkg.cc/apibase/webtype"
)

func validCSRF(c echo.Context, claims any) bool {
	csrfHeader := ""
	switch reflect.TypeOf(claims) {
	case reflect.TypeOf(&t.JwtAccessClaims{}):
		csrfHeader = claims.(*t.JwtAccessClaims).CSRFHeader
	case reflect.TypeOf(&t.JwtRefreshClaims{}):
		csrfHeader = claims.(*t.JwtRefreshClaims).CSRFHeader
	default:
		// c.Logger().Errorf("invalid type for claims")
		return false
	}
	return c.Request().Header.Get("X-XSRF-TOKEN") == csrfHeader
}
