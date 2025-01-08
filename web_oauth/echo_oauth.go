package web_oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/tables"
	"gopkg.cc/apibase/web"
	"gopkg.cc/apibase/web_auth"
)

// Create default routes for oauth user flow
func RegisterOAuthEndpoints(api *web.ApiServer) {
	api.E.GET("/auth/:provider", login(api)) // login & signup
	api.E.GET("/auth/:provider/callback", callback(api))
	api.E.GET("/auth/logout/:provider", logout(api), web_auth.AuthJWT(api))
}

func login(api *web.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		request := c.Request()
		queryURL := request.URL.Query()
		queryURL.Set("provider", c.Param("provider"))
		// TODO: use other way to get referrer that also includes uri fragment (uri including # part)
		referrer := request.Header.Get("referrer")
		// TODO: referrer is a possible attack vector, if it is too large, limit str len to ...

		// Correct Redirecting
		stateBytes, err := json.Marshal(&web.StateReferrer{Nonce: helper.RandomString(16), URI: referrer})
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

func callback(api *web.ApiServer) echo.HandlerFunc {
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

		err = web.JwtLogin(c, api, user, roles)
		if err != nil {
			return err
		}

		// Correct Redirecting
		state := queryURL.Get("state")
		stateBytes, err := base64.StdEncoding.DecodeString(state)
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
		}
		stateReferrer := &web.StateReferrer{}
		err = json.Unmarshal(stateBytes, stateReferrer)
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
		}
		return c.Redirect(http.StatusTemporaryRedirect, stateReferrer.URI)
	}
}

func logout(api *web.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		request := c.Request()
		queryURL := request.URL.Query()
		queryURL.Set("provider", c.Param("provider"))

		gothic.Logout(c.Response(), request) // TODO: check if Logout correctly parses provider from request

		return web.JwtLogout(c, api)
	}
}
