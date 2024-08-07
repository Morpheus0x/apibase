package web

import (
	"fmt"
	"reflect"

	"github.com/labstack/echo/v4"
)

func validCSRF(c echo.Context, claims any) bool {
	fmt.Printf("")
	csrfHeader := ""
	switch reflect.TypeOf(claims) {
	case reflect.TypeOf(&jwtAccessClaims{}):
		csrfHeader = claims.(*jwtAccessClaims).CSRFHeader
	case reflect.TypeOf(&jwtRefreshClaims{}):
		csrfHeader = claims.(*jwtRefreshClaims).CSRFHeader
	default:
		// c.Logger().Errorf("invalid type for claims")
		return false
	}
	return c.Request().Header.Get("X-XSRF-TOKEN") == csrfHeader
}
