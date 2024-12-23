package web_auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	t "gopkg.cc/apibase/webtype"
)

// Create default routes for login and general user flow
func RegisterAuthEndpoints(api *t.ApiServer) error {
	api.E.POST("/auth/login", login(api))
	api.E.POST("/auth/signup", signup(api))
	api.E.GET("/auth/logout", logout(api), AuthJWT(api))

	return nil
}

func login(api *t.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: add fail2ban or similar/more advanced, for login endpoint
		csrfValue := "superRandomCSRF" // TODO: generate randomly

		username := c.FormValue("username")
		password := c.FormValue("password")
		if password != "123456" {
			return echo.ErrUnauthorized
		}
		// TODO: fix jwt role
		accessToken, err := createSignedAccessToken(&t.JwtAccessClaims{Name: username, Role: api.Config.DefaultRole, CSRFHeader: csrfValue}, api)
		if err != nil {
			return fmt.Errorf("unable to create access token: %v", err) // TODO: instead of returning error via http, log it privately on the server
		}
		refreshToken, err := createSignedRefreshToken(&t.JwtRefreshClaims{CSRFHeader: csrfValue}, api)
		if err != nil {
			return fmt.Errorf("unable to create refresh token: %v", err) // TODO: instead of returning error via http, log it privately on the server
		}
		// TODO: save access token to DB
		c.SetCookie(&http.Cookie{Name: "access_token", Value: accessToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(api.Config.TokenAccessValidityDuration() * 2)})
		c.SetCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration() * 2)})
		// TODO: pass csrf token in json response instead of cookie to prevent it from being also sent as cookie on every request
		// csrf token is stored in user jwt, which needs to be parsed for any request anyway
		c.SetCookie(&http.Cookie{Name: "csrf_token", Value: csrfValue, Path: "/", Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration() * 2)})

		return c.JSON(http.StatusOK, map[string]string{"message": "access and refresh token set as cookie"})
	}
}

func signup(api *t.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusNotImplemented, map[string]string{"message": "WIP"})
	}
}

func logout(api *t.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: add option to disable signup or require invite code, add fail2ban
		c.SetCookie(&http.Cookie{Name: "access_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
		c.SetCookie(&http.Cookie{Name: "refresh_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
		// TODO: invalidate refresh_token in DB

		// return c.JSON(http.StatusOK, map[string]string{"message": "Logged out!"})
		return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
	}
}
