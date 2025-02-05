package web

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/errx"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/table"
	wr "gopkg.cc/apibase/web_response"
)

func JwtLogin(c echo.Context, api *ApiServer, user table.User, roles []table.UserRole) error {
	csrfValue := h.RandomString(16) // TODO: protect login page with CSRF, completely separate it from auth jwt
	accessToken, err := createJwtAccessClaims(user.ID, jwtRolesFromTable(roles), user.SuperAdmin, csrfValue).signToken(api)
	if err != nil {
		return wr.NewResponseError(wr.RespErrJwtAccessTokenParsing, errx.Wrapf(err, "unable to create access token for user (id: %d)", user.ID))
	}
	refreshTokenNonce := h.CreateSecretString(h.RandomString(16))
	refreshToken, expiresAt, err := createJwtRefreshClaims(user.ID, refreshTokenNonce, csrfValue).signToken(api)
	if err != nil {
		return wr.NewResponseError(wr.RespErrJwtRefreshTokenParsing, errx.Wrapf(err, "unable to create refresh token for user (id: %d)", user.ID))
	}
	err = api.DB.CreateRefreshTokenEntry(table.RefreshToken{UserID: user.ID, TokenNonce: refreshTokenNonce, ReissueCount: 0, ExpiresAt: expiresAt})
	if err != nil {
		return wr.NewResponseError(wr.RespErrJwtRefreshTokenCreate, errx.Wrapf(err, "unable to create refresh token database entry for user (id: %d)", user.ID))
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
	c.SetCookie(&http.Cookie{Name: "csrf_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})

	refreshToken, err := parseRefreshTokenCookie(c, api.Config.TokenSecretBytes())
	if err != nil {
		return wr.NewResponseError(wr.RespErrJwtRefreshTokenParsing, errx.Wrap(err, "user was logged out but unable to parse refresh token"))
	}
	refreshClaims, ok := refreshToken.Claims.(*jwtRefreshClaims)
	if !ok {
		return wr.NewResponseError(wr.RespErrJwtRefreshTokenClaims, errx.Wrapf(err, "user was logged out but unable to parse refresh claims, refresh token: %v", refreshToken))
	}
	err = api.DB.DeleteRefreshToken(refreshClaims.UserID, refreshClaims.Nonce)
	if err != nil {
		log.Logf(log.LevelError, "user (id: %d) was logged out but unable to delete refresh token (nonce: %s): %s", refreshClaims.UserID, refreshClaims.Nonce, err.Error())
	}
	return nil // default 200 HTTP Status Code // TODO: handle redirect on frontend
}
