package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func authLogin(config ApiConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: add fail2ban
		csrfValue := "superRandomCSRF" // TODO: generate randomly

		username := c.FormValue("username")
		password := c.FormValue("password")
		if password != "123456" {
			return echo.ErrUnauthorized
		}
		accessToken, err := createSignedAccessToken(&jwtAccessClaims{Name: username, Role: SuperAdmin, CSRFHeader: csrfValue}, config)
		if err != nil {
			return fmt.Errorf("unable to create access token: %v", err) // TODO: instead of returning error via http, log it privately on the server
		}
		refreshToken, err := createSignedRefreshToken(&jwtRefreshClaims{Name: username, Role: SuperAdmin, CSRFHeader: csrfValue}, config)
		if err != nil {
			return fmt.Errorf("unable to create refresh token: %v", err) // TODO: instead of returning error via http, log it privately on the server
		}
		// TODO: save access token to DB
		c.SetCookie(&http.Cookie{Name: "access_token", Value: accessToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(config.TokenAccessValidity * 2)})
		c.SetCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(config.TokenRefreshValidity * 2)})
		c.SetCookie(&http.Cookie{Name: "csrf_token", Value: csrfValue, Path: "/", Expires: time.Now().Add(config.TokenRefreshValidity * 2)})

		return c.JSON(http.StatusOK, map[string]string{"message": "access and refresh token set as cookie"})
	}
}

func authLogout(config ApiConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: add option to disable signup or require invite code, add fail2ban
		c.SetCookie(&http.Cookie{Name: "access_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
		c.SetCookie(&http.Cookie{Name: "refresh_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
		// TODO: invalidate refresh_token in DB
		return c.JSON(http.StatusOK, map[string]string{"message": "Logged out!"})
	}
}

func authSignup(config ApiConfig) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusNotImplemented, map[string]string{"message": "WIP"})
	}
}

func authJWT(config ApiConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := authJWTHandler(c, config)
			if err != nil {
				return err
			}
			return next(c)
		}
	}
}

func authJWTHandler(c echo.Context, config ApiConfig) error {
	fmt.Printf("Request Header X-XSRF-TOKEN: %s\n", c.Request().Header.Get("X-XSRF-TOKEN"))
	accessToken, errx := parseAccessTokenCookie(c, config.TokenSecret)
	if errx.IsNil() {
		if !validCSRF(c, accessToken.Claims.(*jwtAccessClaims)) {
			// Invalid CSRF Header received
			return echo.NewHTTPError(http.StatusUnauthorized, "access token CSRF invalid")
		}
		accessTokenExpire, err := accessToken.Claims.GetExpirationTime()
		if accessToken.Valid && err == nil && accessTokenExpire.Time.Add(-time.Minute).After(time.Now()) {
			// Do nothing, access token is still valid for long enough
			return nil
		}
	}
	refreshToken, errx := parseRefreshTokenCookie(c, config.TokenSecret)
	if !errx.IsNil() {
		// c.Logger().Errorf("unable to renew access_token: %s", errx.Text())
		return echo.NewHTTPError(http.StatusUnauthorized, "unable to parse refresh token from cookie")
	}
	if !refreshToken.Valid {
		// c.Logger().Errorf("refresh_token is invalid, unable to renew access_token")
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token invalid")
	}
	// TODO: check DB if refreshToken has been manually invalidated
	// TODO: if refresh token is valid for less than e.g. 1 week, refresh this one also (also refresh csrf_token cookie value and/or expiration)
	refreshClaims, ok := refreshToken.Claims.(*jwtRefreshClaims)
	if !ok {
		// c.Logger().Errorf("unable to parse refresh token claims")
		return echo.NewHTTPError(http.StatusUnauthorized, "unable to parse refresh token claims")
	}
	if !validCSRF(c, refreshToken.Claims.(*jwtRefreshClaims)) {
		// Invalid CSRF Header received
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token CSRF invalid")
	}
	// TODO: get Access Claims from DB
	accessClaims := &jwtAccessClaims{
		Name:       refreshClaims.Name,
		Role:       refreshClaims.Role,
		CSRFHeader: refreshClaims.CSRFHeader,
	}
	newAccessToken, err := createSignedAccessToken(accessClaims, config)
	if err != nil {
		// c.Logger().Errorf("unable to create new access_token")
		return echo.NewHTTPError(http.StatusUnauthorized, "unable to create new access token")
	}
	currentRequest := c.Request()
	// c.Logger().Infof("AllCookies, before adding new access_token: %+v", currentRequest.Cookies())
	newAccessTokenCookie := &http.Cookie{Name: "access_token", Value: newAccessToken, Path: "/", Expires: time.Now().Add(config.TokenAccessValidity * 2)}
	currentRequest.AddCookie(newAccessTokenCookie)
	// c.Logger().Infof("AllCookies, check for duplicate access_token: %+v", currentRequest.Cookies())
	c.SetRequest(currentRequest)
	c.SetCookie(newAccessTokenCookie)
	return nil
}
