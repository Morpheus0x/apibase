package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func login(c echo.Context) error {
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
