package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func authLogin(c echo.Context) error {
	// TODO: Add Dependency Injection for secret and access_token ExpiresAt
	secret := "superSecretSecret"
	accessTokenValidity := time.Second * 15 // time.Hour
	refreshTokenValidity := time.Hour * 24 * 31
	csrfValue := "superRandomCSRF" // TODO: generate randomly

	// log := c.Logger()
	username := c.FormValue("username")
	password := c.FormValue("password")
	if password != "123456" {
		return echo.ErrUnauthorized
	}
	// TODO: use different claims for access and refresh token
	accessToken, err := createSignedAccessToken(&jwtAccessClaims{Name: username, Role: SuperAdmin, CSRFHeader: csrfValue}, accessTokenValidity, secret)
	if err != nil {
		return fmt.Errorf("unable to create access token: %v", err) // TODO: instead of returning error via http, log it privately on the server
	}
	refreshToken, err := createSignedRefreshToken(&jwtRefreshClaims{Name: username, Role: SuperAdmin, CSRFHeader: csrfValue}, refreshTokenValidity, secret)
	if err != nil {
		return fmt.Errorf("unable to create refresh token: %v", err) // TODO: instead of returning error via http, log it privately on the server
	}
	c.SetCookie(&http.Cookie{Name: "access_token", Value: accessToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(accessTokenValidity * 2)})
	c.SetCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(refreshTokenValidity * 2)})
	c.SetCookie(&http.Cookie{Name: "csrf_token", Value: csrfValue, Path: "/", Expires: time.Now().Add(refreshTokenValidity * 2)})
	// return c.JSON(http.StatusOK, echo.Map{
	// 	"accessToken":  accessToken,
	// 	"refreshToken": refreshToken,
	// })
	return c.String(http.StatusOK, "access and refresh token set as cookie")
}

func authLogout(c echo.Context) error {
	c.SetCookie(&http.Cookie{Name: "access_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
	c.SetCookie(&http.Cookie{Name: "refresh_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
	// TODO: invalidate refresh_token in DB
	return c.String(http.StatusOK, "Logged out")
}

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
	// TODO: if refresh token is valid for less than e.g. 1 week, refresh this one also (also refresh csrf_token cookie value and/or expiration)
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
