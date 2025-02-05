package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	wr "gopkg.cc/apibase/web_response"
)

func AuthJWT(api *ApiServer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := authJWTHandler(c, api)
			if c.Response().Committed {
				// Error in JWT Middleware occurred, stop exec but don't return error since response already written
				return nil
			}
			if err != nil {
				return err
			}
			return next(c)
		}
	}
}

func authJWTHandler(c echo.Context, api *ApiServer) error {
	// TODO: use specific error codes for every http error response for easier debugging
	// log.Logf(log.LevelDebug, "Request Header X-XSRF-TOKEN: %s", c.Request().Header.Get("X-XSRF-TOKEN")) // TODO: remove hardcoded header name

	// Verify Access Token
	accessToken, err := parseAccessTokenCookie(c, api.Config.TokenSecretBytes())
	if err == nil {
		accessClaims, ok := accessToken.Claims.(*jwtAccessClaims)
		if ok {
			// TODO: unify the api (error) response using webtype.ApiJsonResponse
			if c.Request().Header.Get("X-XSRF-TOKEN") != accessClaims.CSRFHeader { // TODO: remove hardcoded header name
				// Invalid CSRF Header received
				// log.Logf(log.LevelInfo, "access token CSRF invalid, user: %s, request: %s", accessClaims.Name, c.Request().URL.String())
				return c.JSON(http.StatusForbidden, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrCsrfInvalid})
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
		return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrJwtRefreshTokenParsing})
	}
	if !refreshToken.Valid {
		// log.Logf(log.LevelDebug, "refresh token invalid, request: %s", c.Request().URL.String())
		return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrJwtRefreshTokenInvalid})
	}
	refreshClaims, ok := refreshToken.Claims.(*jwtRefreshClaims)
	if !ok {
		// log.Logf(log.LevelDebug, "unable to parse refresh token claims, request: %s", c.Request().URL.String())
		return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrJwtRefreshTokenClaims})
	}
	if c.Request().Header.Get("X-XSRF-TOKEN") != refreshClaims.CSRFHeader { // TODO: remove hardcoded header name
		// Invalid CSRF Header received
		// log.Logf(log.LevelInfo, "refresh token CSRF invalid, user: %s, request: %s", accessClaims.Name, c.Request().URL.String())
		return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrCsrfInvalid})
	}
	refreshTokenExpire, err := refreshClaims.GetExpirationTime()
	if err != nil || refreshTokenExpire.Time.Before(time.Now()) {
		return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrJwtRefreshTokenExpired})
	}
	valid, err := api.DB.VerifyRefreshTokenNonce(refreshClaims.UserID, refreshClaims.Nonce)
	if err != nil {
		log.Logf(log.LevelDebug, "unable to verify refresh token: %s", err.Error())
		return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrJwtRefreshTokenVerifyErr})
	}
	if !valid {
		return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrJwtRefreshTokenVerifyInvalid})
	}

	// Create New Access Token Claims
	user, err := api.DB.GetUserByID(refreshClaims.UserID)
	if err != nil {
		log.Logf(log.LevelError, "unable to get user from refresh token user id: %s", err.Error())
		return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrUserDoesNotExist})
	}
	roles, err := api.DB.GetUserRoles(refreshClaims.UserID)
	if err != nil {
		log.Logf(log.LevelError, "unable to get roles for jwt access token for user (id: %d): %s", refreshClaims.UserID, err.Error())
		return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrUserNoRoles})
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
	if refreshTokenExpire.Time.Add(-time.Hour * 24 * 7).Before(time.Now()) { // TODO: remove hardcoded timeout
		csrfToken := h.RandomString(16) // TODO: maybe use another random generator...
		newNonce := h.CreateSecretString(h.RandomString(16))

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
			return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrJwtRefreshTokenSigning})
		}
		err = api.DB.UpdateRefreshTokenEntry(refreshClaims.UserID, refreshClaims.Nonce, newNonce, expiresAt)
		if err != nil {
			log.Logf(log.LevelDebug, "unable to update refresh token for user (id: %d): %s", user.ID, err.Error())
			return c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrJwtRefreshTokenUpdate})
		}

		newRefreshTokenCookie := &http.Cookie{Name: "refresh_token", Value: newRefreshToken, Path: "/", Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration() * 2)}
		currentRequest.AddCookie(newRefreshTokenCookie) // set cookie for current request
		c.SetCookie(newRefreshTokenCookie)              // set cookie for response

		// Set new CSRF Cookie, since it was changed with refresh token renew
		c.SetCookie(&http.Cookie{Name: "csrf_token", Value: csrfToken, Path: "/", Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration() * 2)})

		c.Response().Header().Add("X-XSRF-TOKEN", csrfToken) // TODO: remove hardcoded header name
	}

	// Renew Access Token (after check if refresh token sould also be renewed, so that CSRF token will also be updated)
	newAccessToken, err := accessClaims.signToken(api)
	if err != nil {
		log.Logf(log.LevelDebug, "unable to create new access token for user '%s' (id: '%d')", user.Name, user.ID)
		c.JSON(http.StatusUnauthorized, wr.JsonResponse[struct{}]{ErrorID: wr.RespErrJwtAccessTokenSigning})
	}

	newAccessTokenCookie := &http.Cookie{Name: "access_token", Value: newAccessToken, Path: "/", Expires: time.Now().Add(api.Config.TokenAccessValidityDuration() * 2)}
	currentRequest.AddCookie(newAccessTokenCookie) // set cookie for current request
	c.SetCookie(newAccessTokenCookie)              // set cookie for response

	c.SetRequest(currentRequest) // rewrite request with new token(s)
	return nil
}
