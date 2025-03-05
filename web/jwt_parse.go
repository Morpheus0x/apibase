package web

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/errx"
)

func parseAccessTokenCookie[T any](c echo.Context, secret []byte, data T) (*jwt.Token, error) {
	tokenRaw, err := c.Cookie("access_token")
	if err != nil {
		// c.Logger().Debugf("no cookie 'access_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, errx.NewWithType(ErrTokenValidate, "no cookie 'access_token' present in request")
	}
	// this is required in order for Data to be initilized (since new(jwtAccessClaims[T]) doesn't do that)
	accessClaims := &jwtAccessClaims[T]{
		Data: data,
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, accessClaims, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		// c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, errx.NewWithType(ErrTokenValidate, "error parsing token 'access_token'")
	}
	return token, nil
}

func parseRefreshTokenCookie(c echo.Context, secret []byte) (*jwt.Token, error) {
	tokenRaw, err := c.Cookie("refresh_token")
	if err != nil {
		// c.Logger().Debugf("no cookie 'refresh_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, errx.NewWithType(ErrTokenValidate, "no cookie 'refresh_token' present in request")
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(jwtRefreshClaims), func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		// c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, errx.NewWithType(ErrTokenValidate, "error parsing token 'refresh_token'")
	}
	return token, nil
}
