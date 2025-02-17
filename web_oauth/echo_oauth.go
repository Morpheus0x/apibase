package web_oauth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
	"gopkg.cc/apibase/db"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/table"
	"gopkg.cc/apibase/web"
	wr "gopkg.cc/apibase/web_response"
)

// Create default routes for oauth user flow
func RegisterOAuthEndpoints(api *web.ApiServer) {
	api.E.GET("/auth/:provider", login(api), web.CheckCSRF(api)) // login & signup
	api.E.GET("/auth/:provider/callback", callback(api), web.CheckCSRF(api))
	api.E.GET("/auth/logout/:provider", logout(api), web.CheckCSRF(api), web.AuthJWT(api))
}

func login(api *web.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := web.AuthJwtHandler(c, api)
		if err == nil {
			return c.JSON(http.StatusOK, wr.JsonResponse[struct{}]{Message: "already logged in"})
		}

		request := c.Request()
		queryURL := request.URL.Query()
		queryURL.Set("provider", c.Param("provider"))
		referrer := request.Referer()
		if !strings.HasPrefix(referrer, api.Config.AppURI) || len(referrer) > web.REFERRER_MAX_LENGTH {
			referrer = api.Config.AppURI
		}

		// Correct Redirecting
		stateBytes, err := json.Marshal(&web.StateReferrer{Nonce: h.RandomString(16), URI: referrer})
		if err != nil {
			c.Redirect(http.StatusInternalServerError, referrer)
		}
		state := base64.RawURLEncoding.EncodeToString(stateBytes)
		queryURL.Set("state", state)

		// Re-write Request URI w/ provider and state
		request.URL.RawQuery = queryURL.Encode()

		// try to get the user without re-authenticating
		if _, err := gothic.CompleteUserAuth(c.Response(), request); err == nil {
			err := web.AuthJwtHandler(c, api)
			if err == nil {
				return c.Redirect(http.StatusTemporaryRedirect, referrer)
			}
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
			log.Logf(log.LevelError, "oauth callback complete user auth error: %s", err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeyError, wr.RespErrOauthCallbackCompleteAuth).String())
		}
		userToCreate := table.User{
			Name:          gothUser.NickName,
			AuthProvider:  provider,
			Email:         gothUser.Email,
			EmailVerified: false,
		}
		// TODO: impl runSignupDefaultRoleHook to get org assignments from goth.User/userToCreate
		user, err := api.DB.CreateNewUserWithOrg(userToCreate)
		if errors.Is(err, db.ErrOrgCreate) {
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeyError, wr.RespErrSignupNewUserOrg).String())
		}
		if err != nil {
			log.Logf(log.LevelError, "unable to create oauth user db entry: %s", err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeyError, wr.RespErrUserDoesNotExist).String())
		}
		log.Logf(log.LevelDebug, "User logged in: %v", user)

		roles, err := api.DB.GetUserRoles(user.ID)
		if err != nil {
			// roles should already exist or have been created by GetOrCreateUser
			log.Logf(log.LevelError, "unable to get any roles for user (id: %d): %s", user.ID, err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeyError, wr.RespErrUserNoRoles).String())
		}

		newSessionId, err := web.JwtLogin(c, api, user, roles)
		if e, ok := err.(*wr.ResponseError); ok {
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeyError, e.GetErrorId()).String())
		}
		if err != nil {
			log.Logf(log.LevelCritical, "error other than web_response.ResponseError from JwtLogin during oauth callback, this should not happen!: %s", err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeyError, wr.RespErrOauthCallbackUnknownError).String())
		}
		web.UpdateCSRF(c, api, newSessionId)

		// Correct Redirecting
		state := queryURL.Get("state")
		stateBytes, err := base64.RawURLEncoding.DecodeString(state)
		if err != nil {
			log.Logf(log.LevelDevel, "unable to base64 decode state for redirect in oauth callback: %s", err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeySuccess, wr.RespSccsLogin).String())
		}
		stateReferrer := &web.StateReferrer{}
		err = json.Unmarshal(stateBytes, stateReferrer)
		if err != nil {
			log.Logf(log.LevelDevel, "unable to unmarshal StateReferrer json for redirect in oauth callback: %s", err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeySuccess, wr.RespSccsLogin).String())
		}
		if !strings.HasPrefix(stateReferrer.URI, api.Config.AppURI) {
			log.Logf(log.LevelDevel, "StateReferrer URI for redirect doesn't match AppURI(%s) in oauth callback: %s", api.Config.AppURI, stateReferrer.URI)
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeySuccess, wr.RespSccsLogin).String())
		}
		return c.Redirect(http.StatusTemporaryRedirect, stateReferrer.URI)
	}
}

func logout(api *web.ApiServer) echo.HandlerFunc {
	// TODO: maybe not redirect, let that be done by the client?
	return func(c echo.Context) error {
		request := c.Request()
		queryURL := request.URL.Query()
		queryURL.Set("provider", c.Param("provider"))

		err := gothic.Logout(c.Response(), request) // TODO: check if Logout correctly parses provider from request
		if err != nil {
			log.Logf(log.LevelDevel, "error for gothic.Logout() during oauth logout: %s", err.Error())
		}
		web.UpdateCSRF(c, api, h.CreateSecretString(""))
		err = web.JwtLogout(c, api)
		if err, ok := err.(*wr.ResponseError); ok {
			if err.Unwrap() != nil {
				log.Log(log.LevelError, err.Error())
			}
			return c.Redirect(http.StatusInternalServerError, api.Config.AppUri().AddQueryParam(wr.QueryKeyError, err.GetErrorId()).String())
		}
		if err != nil {
			log.Logf(log.LevelCritical, "error other than web_response.ResponseError from JwtLogout during oauth logout, this should not happen!: %s", err.Error())
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeyError, wr.RespErrAuthLogoutUnknownError).String())
		}
		return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppUri().AddQueryParam(wr.QueryKeySuccess, wr.RespSccsLogout).String())
	}
}
