package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/errx"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	wr "gopkg.cc/apibase/web_response"
)

func AuthJWT(api *ApiServer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Log(log.LevelDevel, "Middlware AuthJWT executed")
			err := AuthJwtHandler(c, api)
			if err, ok := err.(*wr.ResponseError); ok {
				if err.Unwrap() != nil {
					log.Log(log.LevelError, err.Error())
				}
				return err.SendJson(c)
			}
			if err != nil {
				// this should not happen, since authJWTHandler always returns web_response.ResponseError
				return err
			}
			return next(c)
		}
	}
}

func AuthJwtHandler(c echo.Context, api *ApiServer) error {
	// Verify Access Token
	accessToken, err := parseAccessTokenCookie(c, api.Config.TokenSecretBytes())
	if err == nil {
		accessClaims, ok := accessToken.Claims.(*jwtAccessClaims)
		if ok {
			accessTokenExpire, err := accessClaims.GetExpirationTime()
			if accessToken.Valid &&
				err == nil &&
				accessTokenExpire.Time.Add(-api.Config.Settings.TokenAccessRenewMargin).After(time.Now()) &&
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
		return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrJwtRefreshTokenParsing, nil)
	}
	refreshClaims, ok := refreshToken.Claims.(*jwtRefreshClaims)
	if !ok {
		// log.Logf(log.LevelDebug, "unable to parse refresh token claims, request: %s", c.Request().URL.String())
		return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrJwtRefreshTokenClaims, nil)
	}
	refreshTokenExpire, err := refreshClaims.GetExpirationTime()
	if err != nil || refreshTokenExpire.Time.Before(time.Now()) {
		return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrJwtRefreshTokenExpired, nil)
	}
	valid, err := api.DB.VerifyRefreshTokenSessionId(refreshClaims.UserID, refreshClaims.SessionID)
	if err != nil {
		log.Logf(log.LevelDebug, "unable to verify refresh token: %s", err.Error())
		return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrJwtRefreshTokenVerifyErr, nil)
	}
	if !valid {
		return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrJwtRefreshTokenVerifyInvalid, nil)
	}

	// Create New Access Token Claims
	user, err := api.DB.GetUserByID(refreshClaims.UserID)
	if err != nil {
		return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrUserDoesNotExist, errx.Wrap(err, "unable to get user from refresh token user id"))
	}
	roles, err := api.DB.GetUserRoles(refreshClaims.UserID)
	if err != nil {
		return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrUserNoRoles, errx.Wrapf(err, "unable to get roles for jwt access token for user (id: %d)", refreshClaims.UserID))
	}
	accessClaims := &jwtAccessClaims{
		UserID:     user.ID,
		Roles:      jwtRolesFromTable(roles),
		SuperAdmin: user.SuperAdmin,
		Revision:   LatestAccessTokenRevision,
	}

	// Get http request to modify it with the new JWTs
	currentRequest := c.Request()

	// Renew Refresh Token, if valid for less than 1 week
	if refreshTokenExpire.Time.Add(-api.Config.Settings.TokenRefreshRenewMargin).Before(time.Now()) {
		newSessionId := h.CreateSecretString(h.RandomString(16)) // TODO: maybe use another random generator...

		newRefreshClaims := &jwtRefreshClaims{
			UserID:    user.ID,
			SessionID: newSessionId,
		}
		newRefreshToken, expiresAt, err := newRefreshClaims.signToken(api)
		if err != nil {
			log.Logf(log.LevelDebug, "unable to create new refresh token for user '%s' (id: '%d')", user.Name, user.ID)
			return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrJwtRefreshTokenSigning, nil)
		}
		userAgent := currentRequest.Header.Get("User-Agent")
		err = api.DB.UpdateRefreshTokenEntry(refreshClaims.UserID, refreshClaims.SessionID, newSessionId, userAgent, expiresAt)
		if err != nil {
			log.Logf(log.LevelDebug, "unable to update refresh token for user (id: %d): %s", user.ID, err.Error())
			return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrJwtRefreshTokenUpdate, nil)
		}

		expiresIn := api.Config.AddCookieExpiryMargin(api.Config.Settings.TokenRefreshValidity)
		newRefreshTokenCookie := &http.Cookie{Name: "refresh_token", Value: newRefreshToken, Path: "/", Expires: time.Now().Add(expiresIn)}

		h.OverwriteRequestCookie(currentRequest, newRefreshTokenCookie) // set cookie for current request
		c.SetCookie(newRefreshTokenCookie)                              // set cookie for response
	}

	// Renew Access Token, since refresh token changed
	newAccessToken, err := accessClaims.signToken(api)
	if err != nil {
		log.Logf(log.LevelDebug, "unable to create new access token for user '%s' (id: '%d')", user.Name, user.ID)
		return wr.NewErrorWithStatus(http.StatusUnauthorized, wr.RespErrJwtAccessTokenSigning, nil)
	}

	expiresIn := api.Config.AddCookieExpiryMargin(api.Config.Settings.TokenAccessValidity)
	newAccessTokenCookie := &http.Cookie{Name: "access_token", Value: newAccessToken, Path: "/", Expires: time.Now().Add(expiresIn)}
	h.OverwriteRequestCookie(currentRequest, newAccessTokenCookie) // set cookie for current request
	c.SetCookie(newAccessTokenCookie)                              // set cookie for response

	c.SetRequest(currentRequest) // rewrite request with new token(s)
	return nil
}
