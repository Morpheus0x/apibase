package web_oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/tables"
	"gopkg.cc/apibase/web_auth"
	t "gopkg.cc/apibase/webtype"
)

// Create default routes for oauth user flow
func RegisterOAuthEndpoints(api *t.ApiServer) error {
	api.E.GET("/auth/:provider", login(api)) // login & signup
	api.E.GET("/auth/:provider/callback", callback(api))
	api.E.GET("/auth/logout/:provider", logout(api), web_auth.AuthJWT(api))

	return nil
}

func login(api *t.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		request := c.Request()
		queryURL := request.URL.Query()
		queryURL.Set("provider", c.Param("provider"))
		// TODO: use other way to get referrer that also includes uri fragment (uri including # part)
		referrer := request.Header.Get("referrer")

		// Correct Redirecting
		stateBytes, err := json.Marshal(&t.StateReferrer{Nonce: helper.RandomString(16), URI: referrer})
		if err != nil {
			c.Redirect(http.StatusInternalServerError, referrer)
		}
		state := base64.StdEncoding.EncodeToString(stateBytes)
		queryURL.Set("state", state)

		// Re-write Request URI w/ provider and state
		request.URL.RawQuery = queryURL.Encode()

		// try to get the user without re-authenticating
		if _, err := gothic.CompleteUserAuth(c.Response(), request); err == nil {
			// TODO: refresh JWT
			return c.Redirect(http.StatusTemporaryRedirect, request.Referer())
		}

		gothic.BeginAuthHandler(c.Response(), request)

		return nil
	}
}

func callback(api *t.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		request := c.Request()
		queryURL := request.URL.Query()
		provider := c.Param("provider")
		queryURL.Set("provider", provider)

		gothUser, err := gothic.CompleteUserAuth(c.Response(), request)
		if err != nil {
			return err
		}
		// TODO: find a way to get org assignments from goth.User
		user, errx := api.Config.DB.GetOrCreateUser(tables.Users{
			Name:          gothUser.NickName,
			AuthProvider:  provider,
			Email:         gothUser.Email,
			EmailVerified: false,
		}, api.Config.DefaultOrgID)
		if !errx.IsNil() {
			errx.Log()
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI) // TODO: add query param or header to show error on client side
		}
		log.Logf(log.LevelDebug, "User logged in: %v", user)

		roles, errx := api.Config.DB.GetUserRoles(user.ID)
		if !errx.IsNil() {
			errx.Extendf("unable to get any roles for user (id: %d)", user.ID).Log()
		}
		csrfValue := helper.RandomString(16) // TODO: protect login page with CSRF, completely separate it from auth jwt
		accessToken, err := t.CreateSignedAccessToken(t.CreateJwtAccessClaims(user.ID, t.JwtRolesFromTable(roles), user.SuperAdmin, csrfValue), api)
		if err != nil {
			log.Logf(log.LevelNotice, "unable to create access token for user (id: %d): %s", user.ID, err.Error())
			return echo.ErrInternalServerError
		}
		refreshTokenNonce := helper.RandomString(16)
		refreshToken, expiresAt, err := t.CreateSignedRefreshToken(t.CreateJwtRefreshClaims(user.ID, refreshTokenNonce, csrfValue), api)
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

		// Correct Redirecting
		state := queryURL.Get("state")
		stateBytes, err := base64.StdEncoding.DecodeString(state)
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
		}
		stateReferrer := &t.StateReferrer{}
		err = json.Unmarshal(stateBytes, stateReferrer)
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
		}
		return c.Redirect(http.StatusTemporaryRedirect, stateReferrer.URI)
	}
}

func logout(api *t.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		request := c.Request()
		queryURL := request.URL.Query()
		queryURL.Set("provider", c.Param("provider"))

		gothic.Logout(c.Response(), request) // TODO: check if Logout correctly parses provider from request

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
		return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
	}
}
