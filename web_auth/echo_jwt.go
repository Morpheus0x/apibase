package web_auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/db"
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
	// TODO: use specific error codes for every http error response for easier debugging
	log.Logf(log.LevelDebug, "Request Header X-XSRF-TOKEN: %s\n", c.Request().Header.Get("X-XSRF-TOKEN"))
	accessToken, errx := parseAccessTokenCookie(c, api.Config.TokenSecret)
	if errx.IsNil() {
		accessClaims, ok := accessToken.Claims.(*t.JwtAccessClaims)
		if ok {
			if !validCSRF(c, accessClaims) {
				// Invalid CSRF Header received
				// log.Logf(log.LevelInfo, "access token CSRF invalid, user: %s, request: %s", accessClaims.Name, c.Request().URL.String())
				return echo.NewHTTPError(http.StatusUnauthorized, "CSRF Error")
			}
			accessTokenExpire, err := accessClaims.GetExpirationTime()
			if accessToken.Valid &&
				err == nil &&
				accessTokenExpire.Time.Add(-time.Minute).After(time.Now()) && // TODO: remove hardcoded timeout
				accessClaims.Revision == t.LatestAccessTokenRevision {
				// Do nothing, access token is still valid for long enough
				return nil
			}
		}
	}
	// TODO: configure client to only send refresh token if access token validity < 2 minute (double of server cutoff)
	// this prevents unnecessary data transmission while still allowing for a single request if refresh token is valid

	// TODO: verify that TokenSecret isn't string converted to bytes, if so TokenSecret must be hex string and decoded to bytes that way
	refreshToken, errx := parseRefreshTokenCookie(c, api.Config.TokenSecret)
	if !errx.IsNil() {
		// log.Logf(log.LevelDebug, "unable to parse refresh token from cookie, request: %s", c.Request().URL.String())
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token cookie")
	}
	if !refreshToken.Valid {
		// log.Logf(log.LevelDebug, "refresh token invalid, request: %s", c.Request().URL.String())
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token invalid")
	}
	refreshClaims, ok := refreshToken.Claims.(*t.JwtRefreshClaims)
	if !ok {
		// log.Logf(log.LevelDebug, "unable to parse refresh token claims, request: %s", c.Request().URL.String())
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token invalid")
	}
	if !validCSRF(c, refreshClaims) {
		// Invalid CSRF Header received
		// log.Logf(log.LevelInfo, "refresh token CSRF invalid, user: %s, request: %s", accessClaims.Name, c.Request().URL.String())
		return echo.NewHTTPError(http.StatusUnauthorized, "CSRF Error")
	}
	valid, errx := db.VerifyRefreshTokenNonce(refreshClaims.UserID, refreshClaims.Nonce)
	if !errx.IsNil() {
		errx.Extend("unable to verify refresh token").Log()
		return echo.NewHTTPError(http.StatusUnauthorized, "internal error, please contact administrator")
	}
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token invalid")
	}

	// Refresh Access Token
	user, errx := db.GetUserByID(refreshClaims.UserID) // TODO: make sure that sql join contains UserRoles[0]
	if !errx.IsNil() {
		errx.Extend("unable to get user from refresh token user id").Log()
		return echo.NewHTTPError(http.StatusUnauthorized, "user doesn't exist")
	}
	accessClaims := &t.JwtAccessClaims{
		UserID: user.ID,
		Name:   user.Name,
		// TODO: fix jwt role
		Role:       api.Config.DefaultRole,
		CSRFHeader: refreshClaims.CSRFHeader,
		Revision:   t.LatestAccessTokenRevision,
	}
	// TODO: check why createSignedAccessToken returns error
	newAccessToken, err := createSignedAccessToken(accessClaims, api)
	if err != nil {
		log.Logf(log.LevelDebug, "unable to create new access token for user '%s' (id: '%s')", user.Name, user.ID)
		// c.Logger().Errorf("unable to create new access_token")
		return echo.NewHTTPError(http.StatusUnauthorized, "internal error, please contact administrator")
	}

	currentRequest := c.Request()

	accessTokenExpire, err := accessClaims.GetExpirationTime()
	// if refresh token is valid for less than 1 week, also refresh it
	if err != nil || accessTokenExpire.Time.Add(-time.Hour*24*7).After(time.Now()) { // TODO: remove hardcoded timeout
		accessClaims.CSRFHeader = "superRandomCSRF" // TODO: generate randomly
		newRefreshClaims := &t.JwtRefreshClaims{
			UserID:     refreshClaims.UserID,
			Nonce:      refreshClaims.Nonce,
			CSRFHeader: "superRandomCSRF", // TODO: generate randomly
		}
		newRefreshToken, err := createSignedRefreshToken(newRefreshClaims, api)
		if err != nil {
			log.Logf(log.LevelDebug, "unable to create new refresh token for user '%s' (id: '%s')", user.Name, user.ID)
			// c.Logger().Errorf("unable to create new access_token")
			return echo.NewHTTPError(http.StatusUnauthorized, "internal error, please contact administrator")
		}
		newRefreshTokenCookie := &http.Cookie{Name: "refresh_token", Value: newRefreshToken, Path: "/", Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration())}
		currentRequest.AddCookie(newRefreshTokenCookie)
		c.SetCookie(newRefreshTokenCookie)
		// TODO: send new CSRF value as Header, so that client doesn't have to parse jwt every time
	}

	newAccessTokenCookie := &http.Cookie{Name: "access_token", Value: newAccessToken, Path: "/", Expires: time.Now().Add(api.Config.TokenAccessValidityDuration())}
	currentRequest.AddCookie(newAccessTokenCookie)
	c.SetCookie(newAccessTokenCookie)

	c.SetRequest(currentRequest)
	return nil
}
