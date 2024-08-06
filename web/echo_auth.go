package web

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/log"
)

func createAndSignToken(claim *jwtClaims, validity time.Duration, secret string) (string, error) {
	now := time.Now()
	claims := &jwtClaims{
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

func refreshAccessToken(c echo.Context) {
	// TODO: Add Dependency Injection for secret and newAccessTokenValidity
	secret := "superSecretSecret"
	newAccessTokenValidity := time.Second * 15 // time.Hour

	// accessTokenIsValid := false
	accessToken, errx := validateToken(c, "access_token", secret)
	if errx.IsNil() {
		// accessTokenIsValid = true
		accessTokenExpire, err := accessToken.Claims.GetExpirationTime()
		if accessToken.Valid && err == nil && accessTokenExpire.Time.Add(-time.Minute).After(time.Now()) {
			// Do nothing, access token is still valid for long enough
			return
		}
	}
	refreshToken, errx := validateToken(c, secret, "refresh_token")
	if !errx.IsNil() {
		c.Logger().Errorf("unable to renew access_token: %s", errx.Text())
		return
	}
	if !refreshToken.Valid {
		c.Logger().Errorf("refresh_token is invalid, unable to renew access_token")
		return
	}
	// TODO: check DB if refreshToken has been manually invalidated
	// TODO: if refresh token is valid for less than e.g. 1 week, refresh this one also
	accessClaims, ok := accessToken.Claims.(*jwtClaims)
	if !ok {
		c.Logger().Errorf("unable to parse claims for new access_token")
		return
	}
	newAccessToken, err := createAndSignToken(accessClaims, newAccessTokenValidity, secret)
	if err != nil {
		c.Logger().Errorf("unable to create new access_token")
		return
	}
	currentRequest := c.Request()
	c.Logger().Infof("AllCookies, before adding new access_token: %+v", currentRequest.Cookies())
	// TODO: maybe use longer cookie Expires time to make sure that the an expired access_token will sent (maybe double validity time)
	newAccessTokenCookie := &http.Cookie{Name: "access_token", Value: newAccessToken, Path: "/", Expires: time.Now().Add(newAccessTokenValidity)}
	currentRequest.AddCookie(newAccessTokenCookie)
	c.Logger().Infof("AllCookies, check for duplicate access_token: %+v", currentRequest.Cookies())
	c.SetRequest(currentRequest)
	c.SetCookie(newAccessTokenCookie)
}

func validateToken(c echo.Context, cookieName string, secret string) (*jwt.Token, log.Err) {
	tokenRaw, err := c.Cookie(cookieName)
	if err != nil {
		c.Logger().Debugf("no cookie '%s' in request (%s): %v", cookieName, c.Request().RequestURI, err)
		return &jwt.Token{}, log.ErrorNew(log.ErrTokenValidate, "no cookie '%s' present in request", cookieName)
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(jwtClaims), func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, log.ErrorNew(log.ErrTokenValidate, "error parsing token '%s'", cookieName)
	}
	return token, log.ErrorNil()
}
