package web_oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"gopkg.cc/apibase/db"
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

// TODO: add jwt logic from local auth
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

		user, err := gothic.CompleteUserAuth(c.Response(), request)
		if err != nil {
			return err
		}
		userFromDB, dbErr := db.GetOrCreateUser(tables.Users{Name: user.NickName, Role: "User", AuthProvider: provider, Email: user.Email, EmailVerified: false})
		if !dbErr.IsNil() {
			dbErr.Log()
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI) // TODO: add query param or header to show error on client side
		}
		log.Logf(log.LevelDebug, "User logged in: %v", userFromDB)
		// TODO: create tokens

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

		// TODO: add option to disable signup or require invite code, add fail2ban
		c.SetCookie(&http.Cookie{Name: "access_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
		c.SetCookie(&http.Cookie{Name: "refresh_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
		// TODO: invalidate refresh_token in DB

		// return c.JSON(http.StatusOK, map[string]string{"message": "Logged out!"})
		return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
	}
}
