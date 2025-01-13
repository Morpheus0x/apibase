package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/table"
)

func JwtLogin(c echo.Context, api *ApiServer, user table.User, roles []table.UserRole) error {
	csrfValue := helper.RandomString(16) // TODO: protect login page with CSRF, completely separate it from auth jwt
	accessToken, err := createJwtAccessClaims(user.ID, jwtRolesFromTable(roles), user.SuperAdmin, csrfValue).signToken(api)
	if err != nil {
		log.Logf(log.LevelNotice, "unable to create access token for user (id: %d): %s", user.ID, err.Error())
		return echo.ErrInternalServerError
	}
	refreshTokenNonce := helper.RandomString(16)
	refreshToken, expiresAt, err := createJwtRefreshClaims(user.ID, refreshTokenNonce, csrfValue).signToken(api)
	if err != nil {
		log.Logf(log.LevelNotice, "unable to create refresh token for user (id: %d): %s", user.ID, err.Error())
		return echo.ErrInternalServerError
	}
	err = api.DB.CreateRefreshTokenEntry(table.RefreshToken{UserID: user.ID, TokenNonce: refreshTokenNonce, ReissueCount: 0, ExpiresAt: expiresAt})
	if err != nil {
		log.Logf(log.LevelNotice, "unable to create refresh token database entry for user (id: %d): %s", user.ID, err.Error())
		return echo.ErrInternalServerError
	}

	// TODO: set cookies to secure/https only (can be configured by ApiConfig setting)
	c.SetCookie(&http.Cookie{Name: "access_token", Value: accessToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(api.Config.TokenAccessValidityDuration() * 2)})
	c.SetCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration() * 2)})
	// TODO: pass csrf token in json response instead of cookie to prevent it from being also sent as cookie on every request
	// csrf token is stored in user jwt, which needs to be parsed for any request anyway
	c.SetCookie(&http.Cookie{Name: "csrf_token", Value: csrfValue, Path: "/", Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration() * 2)})

	return nil
}

func JwtLogout(c echo.Context, api *ApiServer) error {
	c.SetCookie(&http.Cookie{Name: "access_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
	c.SetCookie(&http.Cookie{Name: "refresh_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})

	refreshToken, err := parseRefreshTokenCookie(c, api.Config.TokenSecretBytes())
	if err != nil {
		log.Logf(log.LevelError, "user was logged out but unable to parse refresh token: %s", err.Error())
		return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
	}
	refreshClaims, ok := refreshToken.Claims.(*jwtRefreshClaims)
	if !ok {
		log.Logf(log.LevelError, "user was logged out but unable to parse refresh claims, refresh token: %v: %s", refreshToken, err.Error())
		return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
	}
	err = api.DB.DeleteRefreshToken(refreshClaims.UserID, refreshClaims.Nonce)
	if err != nil {
		log.Logf(log.LevelError, "user (id: %d) was logged out but unable to delete refresh token: %s", refreshClaims.UserID, err.Error())
	}
	// return c.JSON(http.StatusOK, map[string]string{"message": "Logged out!"})
	// TODO: add query param with logout success msg
	return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
}

func AuthJWT(api *ApiServer) echo.MiddlewareFunc {
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

func authJWTHandler(c echo.Context, api *ApiServer) error {
	// TODO: use specific error codes for every http error response for easier debugging
	log.Logf(log.LevelDebug, "Request Header X-XSRF-TOKEN: %s\n", c.Request().Header.Get("X-XSRF-TOKEN")) // TODO: remove hardcoded header name

	// Verify Access Token
	accessToken, err := parseAccessTokenCookie(c, api.Config.TokenSecretBytes())
	if err != nil {
		accessClaims, ok := accessToken.Claims.(*jwtAccessClaims)
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
				accessClaims.Revision == LatestAccessTokenRevision {
				// Do nothing, access token is still valid for long enough
				return nil
			}
		}
	}

	// Verify Refresh Token
	refreshToken, err := parseRefreshTokenCookie(c, api.Config.TokenSecretBytes())
	if err != nil {
		// log.Logf(log.LevelDebug, "unable to parse refresh token from cookie, request: %s", c.Request().URL.String())
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token cookie")
	}
	if !refreshToken.Valid {
		// log.Logf(log.LevelDebug, "refresh token invalid, request: %s", c.Request().URL.String())
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token invalid")
	}
	refreshClaims, ok := refreshToken.Claims.(*jwtRefreshClaims)
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
	valid, err := api.DB.VerifyRefreshTokenNonce(refreshClaims.UserID, refreshClaims.Nonce)
	if err != nil {
		log.Logf(log.LevelDebug, "unable to verify refresh token: %s", err.Error())
		return echo.NewHTTPError(http.StatusUnauthorized, "internal error, please contact administrator")
	}
	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "refresh token invalid")
	}

	// Create New Access Token Claims
	user, err := api.DB.GetUserByID(refreshClaims.UserID)
	if err != nil {
		log.Logf(log.LevelError, "unable to get user from refresh token user id: %s", err.Error())
		return echo.NewHTTPError(http.StatusUnauthorized, "user doesn't exist")
	}
	roles, err := api.DB.GetUserRoles(refreshClaims.UserID)
	if err != nil {
		log.Logf(log.LevelError, "unable to get roles for jwt access token for user (id: %d): %s", refreshClaims.UserID, err.Error())
		return echo.NewHTTPError(http.StatusUnauthorized, "no roles exist for user")
	}
	accessClaims := &jwtAccessClaims{
		UserID:     user.ID,
		Roles:      jwtRolesFromTable(roles),
		SuperAdmin: user.SuperAdmin,
		CSRFHeader: refreshClaims.CSRFHeader,
		Revision:   LatestAccessTokenRevision,
	}

	// Get http request to modify it with the new JWTs
	currentRequest := c.Request()

	// Renew Refresh Token, if valid for less than 1 week
	if refreshTokenExpire.Time.Add(-time.Hour * 24 * 7).After(time.Now()) { // TODO: remove hardcoded timeout
		csrfToken := helper.RandomString(16) // TODO: maybe use another random generator...
		newNonce := helper.RandomString(16)

		// On refresh token renew, also change CSRF token for access token
		accessClaims.CSRFHeader = csrfToken

		newRefreshClaims := &jwtRefreshClaims{
			UserID:     user.ID,
			Nonce:      newNonce,
			CSRFHeader: csrfToken,
		}
		newRefreshToken, expiresAt, err := newRefreshClaims.signToken(api)
		if err != nil {
			log.Logf(log.LevelDebug, "unable to create new refresh token for user '%s' (id: '%d')", user.Name, user.ID)
			return echo.NewHTTPError(http.StatusUnauthorized, "internal error, please contact administrator")
		}
		err = api.DB.UpdateRefreshTokenEntry(refreshClaims.UserID, refreshClaims.Nonce, newNonce, expiresAt)
		if err != nil {
			log.Logf(log.LevelDebug, "unable to update refresh token for user (id: %d): %s", user.ID, err.Error())
			return echo.ErrInternalServerError
		}

		newRefreshTokenCookie := &http.Cookie{Name: "refresh_token", Value: newRefreshToken, Path: "/", Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration())}
		currentRequest.AddCookie(newRefreshTokenCookie) // set cookie for current request
		c.SetCookie(newRefreshTokenCookie)              // set cookie for response

		c.Response().Header().Add("X-XSRF-TOKEN", csrfToken) // TODO: remove hardcoded header name
	}

	// Renew Access Token (after check if refresh token sould also be renewed, so that CSRF token will also be updated)
	newAccessToken, err := accessClaims.signToken(api)
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
