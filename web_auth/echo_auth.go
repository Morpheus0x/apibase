package web_auth

import (
	"errors"
	"net/http"

	"github.com/Morpheus0x/argon2id"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/db"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/table"
	"gopkg.cc/apibase/web"
	wr "gopkg.cc/apibase/web_response"
)

// Create default routes for login and general user flow
func RegisterAuthEndpoints(api *web.ApiServer) {
	api.E.POST("/auth/login", login(api), web.CheckCSRF(api))
	api.E.POST("/auth/signup", signup(api), web.CheckCSRF(api))
	api.E.GET("/auth/logout", logout(api), web.CheckCSRF(api), web.AuthJWT(api))
}

func login(api *web.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: add fail2ban or similar/more advanced, for login endpoint

		err := web.AuthJwtHandler(c, api)
		if err == nil {
			return wr.SendJsonErrorResponse(c, http.StatusOK, wr.RespSccsAlreadyLoggedIn)
		}

		email := c.FormValue("email")
		password := c.FormValue("password")

		if email == "" || password == "" {
			return wr.SendJsonErrorResponse(c, http.StatusUnprocessableEntity, wr.RespErrMissingInput)
		}

		failedHookNr, err := runPreLoginHooks(email, h.CreateSecretString(password))
		if err != nil {
			log.Logf(log.LevelError, "pre login hook %d failed: %s", failedHookNr, err.Error())
			return wr.SendJsonErrorResponse(c, http.StatusInternalServerError, wr.RespErrHookPreLogin)
		}

		user, err := api.DB.GetUserByEmail(email)
		if err != nil {
			log.Logf(log.LevelDebug, "user not found: %s", err.Error())
			return wr.SendJsonErrorResponse(c, http.StatusUnauthorized, wr.RespErrLoginNoUser)
		}

		if user.AuthProvider != "local" {
			return wr.SendJsonErrorResponse(c, http.StatusMisdirectedRequest, wr.RespErrAuthLoginNotLocal)
		}

		match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash.GetSecret())
		if err != nil {
			log.Logf(log.LevelError, "unable to compare password with hash: %s", err.Error())
			return wr.SendJsonErrorResponse(c, http.StatusUnauthorized, wr.RespErrLoginComparePassword)
		}
		if !match {
			return wr.SendJsonErrorResponse(c, http.StatusUnauthorized, wr.RespErrLoginWrongPassword)
		}

		roles, err := api.DB.GetUserRoles(user.ID)
		if err != nil {
			log.Logf(log.LevelError, "no roles exist for user (id: %d), unable to login", user.ID)
			return wr.SendJsonErrorResponse(c, http.StatusUnauthorized, wr.RespErrUserNoRoles)
		}

		failedHookNr, err = runPostLoginHooks(user, roles)
		if err != nil {
			log.Logf(log.LevelError, "post login hook %d failed: %s", failedHookNr, err.Error())
			return wr.SendJsonErrorResponse(c, http.StatusInternalServerError, wr.RespErrHookPostLogin)
		}

		newSessionId, err := web.JwtLogin(c, api, user, roles, api.GetAccessClaimData(user.ID))
		if err, ok := err.(*wr.ResponseError); ok {
			if err.Unwrap() != nil {
				log.Log(log.LevelNotice, err.Error())
			}
			return err.SendJsonWithStatus(c, http.StatusInternalServerError)
		}
		if err != nil {
			log.Logf(log.LevelCritical, "error other than web_response.ResponseError from JwtLogin during login, this should not happen!: %s", err.Error())
			return wr.SendJsonErrorResponse(c, http.StatusInternalServerError, wr.RespErrAuthLoginUnknownError)
		}
		web.UpdateCSRF(c, api, newSessionId)
		return wr.SendJsonErrorResponse(c, http.StatusOK, wr.RespSccsLogin)
	}
}

func signup(api *web.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: add option to disable signup or require invite code, add fail2ban
		email := c.FormValue("email")
		password := c.FormValue("password")
		passwordConfirm := c.FormValue("password-confirm")
		username := c.FormValue("username")

		if email == "" || password == "" || passwordConfirm == "" || username == "" {
			return wr.SendJsonErrorResponse(c, http.StatusUnprocessableEntity, wr.RespErrMissingInput)
		}
		if password != passwordConfirm {
			return wr.SendJsonErrorResponse(c, http.StatusUnprocessableEntity, wr.RespErrSignupPasswordMismatch)
		}

		failedHookNr, err := runPreSignupHooks(username, email, h.CreateSecretString(password))
		if err != nil {
			log.Logf(log.LevelError, "pre signup hook %d failed: %s", failedHookNr, err.Error())
			return wr.SendJsonErrorResponse(c, http.StatusInternalServerError, wr.RespErrHookPreSignup)
		}

		hash, err := argon2id.CreateHash(password, &argonParams)
		if err != nil {
			return wr.SendJsonErrorResponse(c, http.StatusInternalServerError, wr.RespErrSignupPasswordHash)
		}
		userToCreate := table.User{
			Name:           username,
			AuthProvider:   "local",
			Email:          email,
			EmailVerified:  false,
			PasswordHash:   h.CreateSecretString(hash),
			SecretsVersion: 1,
			TotpSecret:     "",
			SuperAdmin:     false,
		}
		rolesToCreate, err := runSignupDefaultRoleHook(userToCreate)
		if err != nil {
			log.Logf(log.LevelError, "signup default hook failed: %s", err.Error())
			return wr.SendJsonErrorResponse(c, http.StatusInternalServerError, wr.RespErrHookSignupDefaultRole)
		}
		user, err := api.DB.CreateNewUserWithOrg(userToCreate, rolesToCreate...)
		if errors.Is(err, db.ErrUserAlreadyExists) {
			return wr.SendJsonErrorResponse(c, http.StatusConflict, wr.RespErrSignupUserExists)
		}
		if errors.Is(err, db.ErrOrgCreate) {
			return wr.SendJsonErrorResponse(c, http.StatusInternalServerError, wr.RespErrSignupNewUserOrg)
		}
		if err != nil {
			return wr.SendJsonErrorResponse(c, http.StatusConflict, wr.RespErrSignupUserCreate)
		}
		roles, err := api.DB.GetUserRoles(user.ID)
		if err != nil {
			// roles should already exist or have been created by CreateUserIfNotExist
			log.Logf(log.LevelError, "unable to get any roles for user (id: %d)", user.ID)
			return wr.SendJsonErrorResponse(c, http.StatusUnauthorized, wr.RespErrUserNoRoles)
		}

		failedHookNr, err = runPostSignupHooks(user, roles)
		if err != nil {
			log.Logf(log.LevelError, "post signup hook %d failed: %s", failedHookNr, err.Error())
		}

		if api.Config.ReqireConfirmEmail {
			// TODO: redirect to page showing email confirmation required,
			// this site as option to input code sent via email,
			// alternatively the user is sent a link containing "id" and "confirmation" query params

			// TODO: Create new email_confirmations table entry with this new id (and secret which will be sent via email as url query)

			confirmId := h.RandomBase64(16)
			return c.JSON(http.StatusOK, wr.JsonResponse[wr.RedirectTarget]{
				ResponseID: wr.RespScssSignupEmailConfirm,
				Data: wr.RedirectTarget{
					// TODO: maybe use another way to get the real referer from the frontend login request
					Referrer: c.Request().Referer(), // TODO: this referer must be used on client side redirect after email code was confirmed (maybe add this also as base64 query param below)
					Target:   api.Config.AppUri().JoinPath("/confirm_email").AddQueryParam("id", confirmId).String(),
				},
			})
		}

		newSessionId, err := web.JwtLogin(c, api, user, roles, api.GetAccessClaimData(user.ID))
		if err, ok := err.(*wr.ResponseError); ok {
			if err.Unwrap() != nil {
				log.Log(log.LevelNotice, err.Error())
			}
			return err.SendJsonWithStatus(c, http.StatusInternalServerError)
		}
		if err != nil {
			log.Logf(log.LevelCritical, "error other than web_response.ResponseError from JwtLogin during signup, this should not happen!: %s", err.Error())
			return wr.SendJsonErrorResponse(c, http.StatusInternalServerError, wr.RespErrAuthSignupUnknownError)
		}
		web.UpdateCSRF(c, api, newSessionId)
		return wr.SendJsonErrorResponse(c, http.StatusOK, wr.RespSccsSignup)
	}
}

func logout(api *web.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		failedHookNr, err := runLogoutHooks(c)
		if err != nil {
			log.Logf(log.LevelError, "logout hook %d failed: %s", failedHookNr, err.Error())
		}
		web.UpdateCSRF(c, api, h.CreateSecretString(""))
		err = web.JwtLogout(c, api)
		if err, ok := err.(*wr.ResponseError); ok {
			if err.Unwrap() != nil {
				log.Log(log.LevelError, err.Error())
			}
			return err.SendJsonWithStatus(c, http.StatusInternalServerError)
		}
		if err != nil {
			log.Logf(log.LevelCritical, "error other than web_response.ResponseError from JwtLogout during logout, this should not happen!: %s", err.Error())
			return wr.SendJsonErrorResponse(c, http.StatusInternalServerError, wr.RespErrAuthLogoutUnknownError)
		}
		return c.JSON(http.StatusOK, wr.JsonResponse[wr.RedirectTarget]{
			ResponseID: wr.RespSccsLogout,
			Data: wr.RedirectTarget{
				Referrer: "none",
				Target:   api.Config.AppUri().AddQueryParam(wr.QueryKeySuccess, wr.RespSccsLogout).String(),
			},
		})
	}
}
