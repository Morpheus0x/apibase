package web_auth

import (
	"fmt"
	"net/http"

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

		err = t.JwtLogin(c, api, user, roles)
		if err != nil {
			return err
		}
		// TODO: fix response
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
		return t.JwtLogout(c, api)
	}
}
