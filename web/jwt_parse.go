package web

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/log"
)

func parseAccessTokenCookie(c echo.Context, secret []byte) (*jwt.Token, *log.Error) {
	tokenRaw, err := c.Cookie("access_token")
	if err != nil {
		// c.Logger().Debugf("no cookie 'access_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "no cookie 'access_token' present in request")
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(jwtAccessClaims), func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		// c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "error parsing token 'access_token'")
	}
	// TODO: add .Valid check for claims?
	return token, log.ErrorNil()
}

func parseRefreshTokenCookie(c echo.Context, secret []byte) (*jwt.Token, *log.Error) {
	tokenRaw, err := c.Cookie("refresh_token")
	if err != nil {
		// c.Logger().Debugf("no cookie 'refresh_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "no cookie 'refresh_token' present in request")
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(jwtRefreshClaims), func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		// c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "error parsing token 'refresh_token'")
	}
	// TODO: add .Valid check for claims?
	return token, log.ErrorNil()
}

func getRefreshClaimsFromCookie(c echo.Context, secret []byte) (jwtRefreshClaims, *log.Error) {
	// TODO: this
	return jwtRefreshClaims{}, log.NewErrorWithType(log.ErrNotImplemented, "WIP")
}
