package hook

import (
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/table"
)

type (
	PreSignupHook  func(username string, email string, password helper.SecretString) error
	PostSignupHook func(user table.User, roles []table.UserRole) error
	PreLoginHook   func(username string, password helper.SecretString) error
	PostLoginHook  func(user table.User, roles []table.UserRole) error
	LogoutHook     func(c echo.Context) error
)

type Hooks struct {
	PreSignup  []PreSignupHook
	PostSignup []PostSignupHook
	PreLogin   []PreLoginHook
	PostLogin  []PostLoginHook
	Logout     []LogoutHook
}

// internal, hooks should be registered with their corresponding functions, not by directly modifying this struct
var RegisteredHooks Hooks

// Prevents signup on error, returns "internal server error", will be logged
func RegisterPreSignupHooks(hooks ...PreSignupHook) {
	RegisteredHooks.PreSignup = append(RegisteredHooks.PreSignup, hooks...)
}

// Runs before email confirmation, doesn't prevent signup on error, will be logged
func RegisterPostSignupHooks(hooks ...PostSignupHook) {
	RegisteredHooks.PostSignup = append(RegisteredHooks.PostSignup, hooks...)
}

// Prevents login on error, returns "internal server error", will be logged
func RegisterPreLoginHooks(hooks ...PreLoginHook) {
	RegisteredHooks.PreLogin = append(RegisteredHooks.PreLogin, hooks...)
}

// Prevents login on error, returns "internal server error", will be logged
func RegisterPostLoginHooks(hooks ...PostLoginHook) {
	RegisteredHooks.PostLogin = append(RegisteredHooks.PostLogin, hooks...)
}

// Doesn't prevent logout on error, will be logged
func RegisterLogoutHooks(hooks ...LogoutHook) {
	RegisteredHooks.Logout = append(RegisteredHooks.Logout, hooks...)
}
