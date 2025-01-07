package web_auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/helper"
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
	log.Logf(log.LevelDebug, "Request Header X-XSRF-TOKEN: %s\n", c.Request().Header.Get("X-XSRF-TOKEN")) // TODO: remove hardcoded header name

	// Verify Access Token
	accessToken, errx := t.ParseAccessTokenCookie(c, api.Config.TokenSecretBytes())
	if errx.IsNil() {
		accessClaims, ok := accessToken.Claims.(*t.JwtAccessClaims)
		if ok {
			// TODO: unify the api (error) response using webtype.ApiJsonResponse
			if c.Request().Header.Get("X-XSRF-TOKEN") != accessClaims.CSRFHeader { // TODO: remove hardcoded header name
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

	// Verify Refresh Token
	refreshToken, errx := t.ParseRefreshTokenCookie(c, api.Config.TokenSecretBytes())
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
	if c.Request().Header.Get("X-XSRF-TOKEN") != refreshClaims.CSRFHeader { // TODO: remove hardcoded header name
		// Invalid CSRF Header received
		// log.Logf(log.LevelInfo, "refresh token CSRF invalid, user: %s, request: %s", accessClaims.Name, c.Request().URL.String())
		return echo.NewHTTPError(http.StatusUnauthorized, "CSRF Error")
	}
	refreshTokenExpire, err := refreshClaims.GetExpirationTime()
	if err != nil || refreshTokenExpire.Time.After(time.Now()) {
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token expired")
	}
	valid, errx := api.Config.DB.VerifyRefreshTokenNonce(refreshClaims.UserID, refreshClaims.Nonce)
	if !errx.IsNil() {
		errx.Extend("unable to verify refresh token").Severity(log.LevelDebug).Log()
		return echo.NewHTTPError(http.StatusUnauthorized, "internal error, please contact administrator")
	}
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token invalid")
	}

	// Create New Access Token Claims
	user, errx := api.Config.DB.GetUserByID(refreshClaims.UserID) // TODO: make sure that sql join contains UserRoles[0]
	if !errx.IsNil() {
		errx.Extend("unable to get user from refresh token user id").Log()
		return echo.NewHTTPError(http.StatusUnauthorized, "user doesn't exist")
	}
	roles, errx := api.Config.DB.GetUserRoles(refreshClaims.UserID)
	if !errx.IsNil() {
		errx.Extendf("unable to get roles for jwt access token for user (id: %d)", refreshClaims.UserID).Log()
	}
	accessClaims := &t.JwtAccessClaims{
		UserID:     user.ID,
		Roles:      t.JwtRolesFromTable(roles),
		SuperAdmin: user.SuperAdmin,
		CSRFHeader: refreshClaims.CSRFHeader,
		Revision:   t.LatestAccessTokenRevision,
	}

	// Get http request to modify it with the new JWTs
	currentRequest := c.Request()

	// Renew Refresh Token, if valid for less than 1 week
	if refreshTokenExpire.Time.Add(-time.Hour * 24 * 7).After(time.Now()) { // TODO: remove hardcoded timeout
		csrfToken := helper.RandomString(16) // TODO: maybe use another random generator...
		newNonce := helper.RandomString(16)

		// On refresh token renew, also change CSRF token for access token
		accessClaims.CSRFHeader = csrfToken

		newRefreshClaims := &t.JwtRefreshClaims{
			UserID:     user.ID,
			Nonce:      newNonce,
			CSRFHeader: csrfToken,
		}
		newRefreshToken, expiresAt, err := t.CreateSignedRefreshToken(newRefreshClaims, api)
		if err != nil {
			log.Logf(log.LevelDebug, "unable to create new refresh token for user '%s' (id: '%d')", user.Name, user.ID)
			// c.Logger().Errorf("unable to create new access_token")
			return echo.NewHTTPError(http.StatusUnauthorized, "internal error, please contact administrator")
		}
		errx := api.Config.DB.UpdateRefreshTokenEntry(refreshClaims.UserID, refreshClaims.Nonce, newNonce, expiresAt)
		if !errx.IsNil() {
			errx.Extendf("unable to update refresh token for user (id: %d)", user.ID).Severity(log.LevelDebug).Log()
			return echo.ErrInternalServerError
		}

		newRefreshTokenCookie := &http.Cookie{Name: "refresh_token", Value: newRefreshToken, Path: "/", Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration())}
		currentRequest.AddCookie(newRefreshTokenCookie) // set cookie for current request
		c.SetCookie(newRefreshTokenCookie)              // set cookie for response

		c.Response().Header().Add("X-XSRF-TOKEN", csrfToken) // TODO: remove hardcoded header name
	}

	// Renew Access Token (after check if refresh token sould also be renewed, so that CSRF token will also be updated)
	newAccessToken, err := t.CreateSignedAccessToken(accessClaims, api)
	if err != nil {
		log.Logf(log.LevelDebug, "unable to create new access token for user '%s' (id: '%d')", user.Name, user.ID)
		// c.Logger().Errorf("unable to create new access_token")
		return echo.NewHTTPError(http.StatusUnauthorized, "internal error, please contact administrator")
	}

	newAccessTokenCookie := &http.Cookie{Name: "access_token", Value: newAccessToken, Path: "/", Expires: time.Now().Add(api.Config.TokenAccessValidityDuration())}
	currentRequest.AddCookie(newAccessTokenCookie) // set cookie for current request
	c.SetCookie(newAccessTokenCookie)              // set cookie for response

	c.SetRequest(currentRequest) // rewrite request with new token(s)
	return nil
}
