package web_auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Morpheus0x/argon2id"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/tables"
	t "gopkg.cc/apibase/webtype"
)

// Create default routes for login and general user flow
func RegisterAuthEndpoints(api *t.ApiServer) {
	api.E.POST("/auth/login", login(api))
	api.E.POST("/auth/signup", signup(api))
	api.E.GET("/auth/logout", logout(api), AuthJWT(api))
}

func login(api *t.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: add fail2ban or similar/more advanced, for login endpoint

		email := c.FormValue("email")
		password := c.FormValue("password")

		user, errx := api.Config.DB.GetUserByEmail(email)
		if !errx.IsNil() {
			return fmt.Errorf("user not found: %s", errx.String())
		}

		// TODO: unify the api (error) response using webtype.ApiJsonResponse
		match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash)
		if err != nil {
			log.Logf(log.LevelError, "unable to compare password with hash: %s", err.Error())
			return echo.ErrUnauthorized // c.String(http.StatusUnauthorized, "unable to verify password")
		}
		if !match {
			return echo.ErrUnauthorized // c.String(http.StatusUnauthorized, "invalid password")
		}

		roles, errx := api.Config.DB.GetUserRoles(user.ID)
		if !errx.IsNil() {
			errx.Extendf("unable to get any roles for user (id: %d)", user.ID).Log()
		}

		csrfValue := helper.RandomString(16) // TODO: protect login page with CSRF, completely separate it from auth jwt
		accessToken, err := t.CreateJwtAccessClaims(user.ID, t.JwtRolesFromTable(roles), user.SuperAdmin, csrfValue).SignToken(api)
		if err != nil {
			log.Logf(log.LevelNotice, "unable to create access token for user (id: %d): %s", user.ID, err.Error())
			return echo.ErrInternalServerError
		}
		refreshTokenNonce := helper.RandomString(16)
		refreshToken, expiresAt, err := t.CreateJwtRefreshClaims(user.ID, refreshTokenNonce, csrfValue).SignToken(api)
		if err != nil {
			log.Logf(log.LevelNotice, "unable to create refresh token for user (id: %d): %s", user.ID, err.Error())
			return echo.ErrInternalServerError
		}
		errx = api.Config.DB.CreateRefreshTokenEntry(tables.RefreshTokens{UserID: user.ID, TokenNonce: refreshTokenNonce, ReissueCount: 0, ExpiresAt: expiresAt})
		if !errx.IsNil() {
			log.Logf(log.LevelNotice, "unable to create refresh token database entry for user (id: %d): %s", user.ID, err.Error())
			return echo.ErrInternalServerError
		}
		// TODO: set cookies to secure/https only (can be configured by ApiConfig setting)
		c.SetCookie(&http.Cookie{Name: "access_token", Value: accessToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(api.Config.TokenAccessValidityDuration() * 2)})
		c.SetCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken, Path: "/", HttpOnly: true, Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration() * 2)})
		// TODO: pass csrf token in json response instead of cookie to prevent it from being also sent as cookie on every request
		// csrf token is stored in user jwt, which needs to be parsed for any request anyway
		c.SetCookie(&http.Cookie{Name: "csrf_token", Value: csrfValue, Path: "/", Expires: time.Now().Add(api.Config.TokenRefreshValidityDuration() * 2)})

		// TODO: Correct Redirecting
		return c.JSON(http.StatusOK, map[string]string{"message": "access and refresh token set as cookie"})
	}
}

func signup(api *t.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: add option to disable signup or require invite code, add fail2ban

		// TODO: this
		return c.JSON(http.StatusNotImplemented, map[string]string{"message": "WIP"})
	}
}

func logout(api *t.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.SetCookie(&http.Cookie{Name: "access_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
		c.SetCookie(&http.Cookie{Name: "refresh_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})

		refreshToken, errx := t.ParseRefreshTokenCookie(c, api.Config.TokenSecretBytes())
		if !errx.IsNil() {
			errx.Extendf("user was logged out but unable to parse refresh token").Log()
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
		}
		refreshClaims, ok := refreshToken.Claims.(*t.JwtRefreshClaims)
		if !ok {
			errx.Extendf("user was logged out but unable to parse refresh claims, refresh token: %v", refreshToken).Log()
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
		}
		errx = api.Config.DB.DeleteRefreshToken(refreshClaims.UserID, refreshClaims.Nonce)
		if !errx.IsNil() {
			errx.Extendf("user (id: %d) was logged out but unable to delete refresh token", refreshClaims.UserID).Log()
		}
		// return c.JSON(http.StatusOK, map[string]string{"message": "Logged out!"})
		// TODO: add query param with logout success msg
		return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
	}
}
