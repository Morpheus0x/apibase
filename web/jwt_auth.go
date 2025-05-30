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

func JwtLogin(c echo.Context, api *ApiServer, user table.User, roles []table.UserRole, accessClaimData any) (h.SecretString, error) {
	noNewSession := h.CreateSecretString("")
	newSessionId := h.CreateSecretString(h.RandomBase64(32))
	accessToken, err := CreateJwtAccessClaims(user.ID, jwtRolesFromTable(roles), user.SuperAdmin, accessClaimData).SignToken(api)
	if err != nil {
		return noNewSession, wr.NewError(wr.RespErrJwtAccessTokenParsing, errx.Wrapf(err, "unable to create access token for user (id: %d)", user.ID))
	}
	refreshToken, expiresAt, err := createJwtRefreshClaims(user.ID, newSessionId).signToken(api)
	if err != nil {
		return noNewSession, wr.NewError(wr.RespErrJwtRefreshTokenParsing, errx.Wrapf(err, "unable to create refresh token for user (id: %d)", user.ID))
	}
	userAgent := c.Request().Header.Get("User-Agent")
	err = api.DB.CreateRefreshTokenEntry(table.RefreshToken{UserID: user.ID, SessionID: newSessionId, ReissueCount: 0, UserAgent: userAgent, ExpiresAt: expiresAt})
	if err != nil {
		return noNewSession, wr.NewError(wr.RespErrJwtRefreshTokenCreate, errx.Wrapf(err, "unable to create refresh token database entry for user (id: %d)", user.ID))
	}

	expiresIn := api.Config.AddCookieExpiryMargin(api.Config.Settings.TokenAccessValidity)
	c.SetCookie(&http.Cookie{Name: "access_token", Value: accessToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(expiresIn)})
	expiresIn = api.Config.AddCookieExpiryMargin(api.Config.Settings.TokenRefreshValidity)
	c.SetCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(expiresIn)})

	return newSessionId, nil
}

func JwtLogout(c echo.Context, api *ApiServer) error {
	c.SetCookie(&http.Cookie{Name: "access_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
	c.SetCookie(&http.Cookie{Name: "refresh_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})

	refreshToken, err := parseRefreshTokenCookie(c, api.Config.TokenSecretBytes())
	if err != nil {
		return wr.NewError(wr.RespErrJwtRefreshTokenParsing, errx.Wrap(err, "user was logged out but unable to parse refresh token"))
	}
	refreshClaims, ok := refreshToken.Claims.(*jwtRefreshClaims)
	if !ok {
		return wr.NewError(wr.RespErrJwtRefreshTokenClaims, errx.Wrapf(err, "user was logged out but unable to parse refresh claims, refresh token: %v", refreshToken))
	}
	err = api.DB.DeleteRefreshToken(refreshClaims.UserID, refreshClaims.SessionID)
	if err != nil {
		log.Logf(log.LevelError, "user (id: %d) was logged out but unable to delete refresh token (session id: %s): %s", refreshClaims.UserID, refreshClaims.SessionID, err.Error())
	}
	return nil
}
