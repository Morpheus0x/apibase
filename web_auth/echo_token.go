package web_auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/config"
	"gopkg.cc/apibase/log"
	t "gopkg.cc/apibase/webtype"
)

func createSignedAccessToken(claims *t.JwtAccessClaims, api *config.ApiServer) (string, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(api.Config.TokenAccessValidity))
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return rawToken.SignedString([]byte(api.Config.TokenSecret))
}

func createSignedRefreshToken(claims *t.JwtRefreshClaims, api *config.ApiServer) (string, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(api.Config.TokenRefreshValidity))
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return rawToken.SignedString([]byte(api.Config.TokenSecret))
}

func parseAccessTokenCookie(c echo.Context, secret string) (*jwt.Token, *log.Error) {
	tokenRaw, err := c.Cookie("access_token")
	if err != nil {
		// c.Logger().Debugf("no cookie 'access_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "no cookie 'access_token' present in request")
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(t.JwtAccessClaims), func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		// c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "error parsing token 'access_token'")
	}
	return token, log.ErrorNil()
}

func parseRefreshTokenCookie(c echo.Context, secret string) (*jwt.Token, *log.Error) {
	tokenRaw, err := c.Cookie("refresh_token")
	if err != nil {
		// c.Logger().Debugf("no cookie 'refresh_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "no cookie 'refresh_token' present in request")
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(t.JwtRefreshClaims), func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		// c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "error parsing token 'refresh_token'")
	}
	return token, log.ErrorNil()
}
