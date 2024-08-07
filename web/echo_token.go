package web

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/log"
)

func createSignedAccessToken(claim *jwtAccessClaims, validity time.Duration, secret string) (string, error) {
	now := time.Now()
	claims := &jwtAccessClaims{
		claim.Name,
		claim.Role,
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(validity)),
		},
	}
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return rawToken.SignedString([]byte(secret))
}

func createSignedRefreshToken(claim *jwtRefreshClaims, validity time.Duration, secret string) (string, error) {
	now := time.Now()
	claims := &jwtRefreshClaims{
		claim.Name,
		claim.Role,
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(validity)),
		},
	}
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return rawToken.SignedString([]byte(secret))
}

func parseAccessTokenCookie(c echo.Context, secret string) (*jwt.Token, log.Err) {
	tokenRaw, err := c.Cookie("access_token")
	if err != nil {
		c.Logger().Debugf("no cookie 'access_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, log.ErrorNew(log.ErrTokenValidate, "no cookie 'access_token' present in request")
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(jwtAccessClaims), func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, log.ErrorNew(log.ErrTokenValidate, "error parsing token 'access_token'")
	}
	return token, log.ErrorNil()
}

func parseRefreshTokenCookie(c echo.Context, secret string) (*jwt.Token, log.Err) {
	tokenRaw, err := c.Cookie("refresh_token")
	if err != nil {
		c.Logger().Debugf("no cookie 'refresh_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, log.ErrorNew(log.ErrTokenValidate, "no cookie 'refresh_token' present in request")
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(jwtRefreshClaims), func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, log.ErrorNew(log.ErrTokenValidate, "error parsing token 'refresh_token'")
	}
	return token, log.ErrorNil()
}
