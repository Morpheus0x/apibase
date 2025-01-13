package web_auth

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Morpheus0x/argon2id"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/db"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/table"
	"gopkg.cc/apibase/web"
)

// Create default routes for login and general user flow
func RegisterAuthEndpoints(api *web.ApiServer) {
	api.E.POST("/auth/login", login(api))
	api.E.POST("/auth/signup", signup(api))
	api.E.GET("/auth/logout", logout(api), web.AuthJWT(api))
}

func login(api *web.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		// TODO: add fail2ban or similar/more advanced, for login endpoint

		email := c.FormValue("email")
		password := c.FormValue("password")

		user, err := api.DB.GetUserByEmail(email)
		if err != nil {
			log.Logf(log.LevelDebug, "user not found: %s", err.Error())
			return c.JSON(http.StatusUnauthorized, "user doesn't exist")
		}

		// TODO: unify the api (error) response using webtype.ApiJsonResponse
		match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash.GetSecret())
		if err != nil {
			log.Logf(log.LevelError, "unable to compare password with hash: %s", err.Error())
			return echo.ErrUnauthorized // c.String(http.StatusUnauthorized, "unable to verify password")
		}
		if !match {
			return echo.ErrUnauthorized // c.String(http.StatusUnauthorized, "invalid password")
		}

		roles, err := api.DB.GetUserRoles(user.ID)
		if err != nil {
			log.Logf(log.LevelError, "no roles exist for user (id: %d), unable to login", user.ID)
			return c.JSON(http.StatusUnauthorized, map[string]string{"message": "no roles exist for user"})
		}

		err = web.JwtLogin(c, api, user, roles)
		if err != nil {
			return err
		}
		// TODO: fix response
		return c.JSON(http.StatusOK, map[string]string{"message": "access and refresh token set as cookie"})
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
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"message": "missing input"})
		}
		if password != passwordConfirm {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"message": "password confirmation doesn't match"})
		}
		hash, err := argon2id.CreateHash(password, &argonParams)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
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
		role := web.DefaultRole
		if r, ok := api.Config.DefaultOrgRole[strconv.Itoa(api.Config.DefaultOrgID)]; ok {
			role = r
		}
		user, err := api.DB.CreateUserIfNotExist(userToCreate, role.GetTable(0, api.Config.DefaultOrgID))
		if errors.Is(err, db.ErrUserAlreadyExists) {
			return c.JSON(http.StatusConflict, map[string]string{"message": "user with that email already exists"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
		}
		roles, err := api.DB.GetUserRoles(user.ID)
		if err != nil {
			// roles should already exist or have been created by CreateUserIfNotExist
			log.Logf(log.LevelError, "unable to get any roles for user (id: %d)", user.ID)
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
		}

		if api.Config.ReqireConfirmEmail {
			// TODO: redirect to page showing email confirmation required,
			// have option to input code sent via email to directly redirect to where user left off
			// for this to work the redirect target must be passed to that page via ... header?
			return c.Redirect(http.StatusTemporaryRedirect, api.Config.AppURI)
		}

		err = web.JwtLogin(c, api, user, roles)
		if err != nil {
			return err
		}
		// TODO: fix response
		return c.JSON(http.StatusOK, map[string]string{"message": "access and refresh token set as cookie"})
	}
}

func logout(api *web.ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		return web.JwtLogout(c, api)
	}
}
