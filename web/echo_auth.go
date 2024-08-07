package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func authMiddlewareWrapper(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := authMiddleware(c)
		if err != nil {
			return err
		}
		return next(c)
	}
}

func authMiddleware(c echo.Context) error {
	// TODO: Add Dependency Injection for secret and newAccessTokenValidity
	secret := "superSecretSecret"
	newAccessTokenValidity := time.Second * 15 // time.Hour

	accessToken, errx := parseAccessTokenCookie(c, secret)
	if errx.IsNil() {
		if csrfInvalid(c, accessToken.Claims.(*jwtAccessClaims)) {
			// Invalid CSRF Header received
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}
		accessTokenExpire, err := accessToken.Claims.GetExpirationTime()
		if accessToken.Valid && err == nil && accessTokenExpire.Time.Add(-time.Minute).After(time.Now()) {
			// Do nothing, access token is still valid for long enough
			return nil
		}
	}
	refreshToken, errx := parseRefreshTokenCookie(c, secret)
	if !errx.IsNil() {
		// c.Logger().Errorf("unable to renew access_token: %s", errx.Text())
		return nil
	}
	if !refreshToken.Valid {
		// c.Logger().Errorf("refresh_token is invalid, unable to renew access_token")
		return nil
	}
	// TODO: check DB if refreshToken has been manually invalidated
	// TODO: if refresh token is valid for less than e.g. 1 week, refresh this one also
	refreshClaims, ok := refreshToken.Claims.(*jwtRefreshClaims)
	if !ok {
		// c.Logger().Errorf("unable to parse refresh token claims")
		return nil
	}
	if csrfInvalid(c, refreshToken.Claims.(*jwtRefreshClaims)) {
		// Invalid CSRF Header received
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}
	// TODO: get Access Claims from DB
	accessClaims := &jwtAccessClaims{
		Name:       refreshClaims.Name,
		Role:       refreshClaims.Role,
		CSRFHeader: refreshClaims.CSRFHeader,
	}
	newAccessToken, err := createSignedAccessToken(accessClaims, newAccessTokenValidity, secret)
	if err != nil {
		// c.Logger().Errorf("unable to create new access_token")
		return nil
	}
	currentRequest := c.Request()
	// c.Logger().Infof("AllCookies, before adding new access_token: %+v", currentRequest.Cookies())
	newAccessTokenCookie := &http.Cookie{Name: "access_token", Value: newAccessToken, Path: "/", Expires: time.Now().Add(newAccessTokenValidity * 2)}
	currentRequest.AddCookie(newAccessTokenCookie)
	// c.Logger().Infof("AllCookies, check for duplicate access_token: %+v", currentRequest.Cookies())
	c.SetRequest(currentRequest)
	c.SetCookie(newAccessTokenCookie)
	return nil
}
