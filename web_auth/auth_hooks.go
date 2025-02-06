package web_auth

import (
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/hook"
	"gopkg.cc/apibase/table"
)

func runPreSignupHooks(username string, email string, password helper.SecretString) (int, error) {
	var err error
	for failedHookNr, hook := range hook.RegisteredHooks.PreSignup {
		err = hook(username, email, password)
		if err != nil {
			return failedHookNr + 1, err
		}
	}
	return 0, nil
}

func runSignupDefaultRoleHook(user table.User) ([]table.UserRole, error) {
	if len(hook.RegisteredHooks.SignupDefaultRole) < 1 {
		return []table.UserRole{}, nil
	}
	return hook.RegisteredHooks.SignupDefaultRole[0](user)
}

func runPostSignupHooks(user table.User, roles []table.UserRole) (int, error) {
	var err error
	for failedHookNr, hook := range hook.RegisteredHooks.PostSignup {
		err = hook(user, roles)
		if err != nil {
			return failedHookNr + 1, err
		}
	}
	return 0, nil
}

func runPreLoginHooks(username string, password helper.SecretString) (int, error) {
	var err error
	for failedHookNr, hook := range hook.RegisteredHooks.PreLogin {
		err = hook(username, password)
		if err != nil {
			return failedHookNr + 1, err
		}
	}
	return 0, nil
}

func runPostLoginHooks(user table.User, roles []table.UserRole) (int, error) {
	var err error
	for failedHookNr, hook := range hook.RegisteredHooks.PostLogin {
		err = hook(user, roles)
		if err != nil {
			return failedHookNr + 1, err
		}
	}
	return 0, nil
}

func runLogoutHooks(c echo.Context) (int, error) {
	var err error
	for failedHookNr, hook := range hook.RegisteredHooks.Logout {
		err = hook(c)
		if err != nil {
			return failedHookNr + 1, err
		}
	}
	return 0, nil
}
