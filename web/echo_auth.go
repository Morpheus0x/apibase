package web

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
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

	accessTokenRaw, err := c.Cookie("access_token")
	if err != nil {
		// ###################
		// TODO: Setting a very low cookie expiration makes it so that the request doesn't even send the access_token
		// Therefore, either set access_token cookie expiration to same as refresh_token or don't require an access_token here
		// ###################
		c.Logger().Error("Do nothing, since auth will fail, since no access_token cookie was provided") // TODO: remove
		// Do nothing, since auth will fail, since no access_token cookie was provided
		return
	}
	accessToken, err := jwt.ParseWithClaims(accessTokenRaw.Value, new(jwtClaims), func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		c.Logger().Errorf("Do nothing, since auth will fail, since access_token cannot be parsed, err: %+v", err) // TODO: remove
		// Do nothing, since auth will fail, since access_token cannot be parsed
		return
	}
	accessTokenExpire, err := accessToken.Claims.GetExpirationTime()
	if accessToken.Valid && err == nil && accessTokenExpire.Time.Add(-time.Minute).After(time.Now()) {
		c.Logger().Errorf("Do nothing, access token is still valid for long enough") // TODO: remove
		// Do nothing, access token is still valid for long enough
		return
	}
	refreshTokenRaw, err := c.Cookie("refresh_token")
	if err != nil {
		c.Logger().Errorf("no cookie refresh_token, unable to renew access_token")
		return
	}
	refreshToken, err := jwt.ParseWithClaims(refreshTokenRaw.Value, new(jwtClaims), func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !refreshToken.Valid {
		c.Logger().Errorf("refresh_token is invalid, unable to renew access_token")
		return
	}
	// TODO: check DB if refreshToken has been manually invalidated
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
	newAccessTokenCookie := &http.Cookie{Name: "access_token", Value: newAccessToken, Path: "/", Expires: time.Now().Add(newAccessTokenValidity)}
	currentRequest.AddCookie(newAccessTokenCookie)
	c.Logger().Infof("AllCookies, check for duplicate access_token: %+v", currentRequest.Cookies())
	c.SetRequest(currentRequest)
	c.SetCookie(newAccessTokenCookie)
}
