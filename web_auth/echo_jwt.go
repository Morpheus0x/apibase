package web_auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/log"
	t "gopkg.cc/apibase/webtype"
)

func AuthJWT(api *t.ApiServer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := authJWTHandler(c, api)
			if err != nil {
				return err
			}
			return next(c)
		}
	}
}

func authJWTHandler(c echo.Context, api *t.ApiServer) error {
	log.Logf(log.LevelDebug, "Request Header X-XSRF-TOKEN: %s\n", c.Request().Header.Get("X-XSRF-TOKEN"))
	accessToken, errx := parseAccessTokenCookie(c, api.Config.TokenSecret)
	if errx.IsNil() {
		if !validCSRF(c, accessToken.Claims.(*t.JwtAccessClaims)) {
			// Invalid CSRF Header received
			return echo.NewHTTPError(http.StatusUnauthorized, "access token CSRF invalid")
		}
		accessTokenExpire, err := accessToken.Claims.GetExpirationTime()
		if accessToken.Valid && err == nil && accessTokenExpire.Time.Add(-time.Minute).After(time.Now()) {
			// Do nothing, access token is still valid for long enough
			return nil
		}
	}
	// TODO: configure client to only send refresh token if access token validity < 2 minute (double of server cutoff)
	// this prevents unnecessary data transmission while still allowing for a single request if refresh token is valid
	refreshToken, errx := parseRefreshTokenCookie(c, api.Config.TokenSecret)
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
	refreshClaims, ok := refreshToken.Claims.(*t.JwtRefreshClaims)
	if !ok {
		// c.Logger().Errorf("unable to parse refresh token claims")
		return echo.NewHTTPError(http.StatusUnauthorized, "unable to parse refresh token claims")
	}
	if !validCSRF(c, refreshToken.Claims.(*t.JwtRefreshClaims)) {
		// Invalid CSRF Header received
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token CSRF invalid")
	}
	// TODO: get Access Claims from DB
	accessClaims := &t.JwtAccessClaims{
		Name:       refreshClaims.Name,
		Role:       refreshClaims.Role,
		CSRFHeader: refreshClaims.CSRFHeader,
	}
	newAccessToken, err := createSignedAccessToken(accessClaims, api)
	if err != nil {
		// c.Logger().Errorf("unable to create new access_token")
		return echo.NewHTTPError(http.StatusUnauthorized, "unable to create new access token")
	}
	currentRequest := c.Request()
	// c.Logger().Infof("AllCookies, before adding new access_token: %+v", currentRequest.Cookies())
	newAccessTokenCookie := &http.Cookie{Name: "access_token", Value: newAccessToken, Path: "/", Expires: time.Now().Add(api.Config.TokenAccessValidityDuration() * 2)}
	currentRequest.AddCookie(newAccessTokenCookie)
	// c.Logger().Infof("AllCookies, check for duplicate access_token: %+v", currentRequest.Cookies())
	c.SetRequest(currentRequest)
	c.SetCookie(newAccessTokenCookie)
	return nil
}
