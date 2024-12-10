package config

import (
	"math/rand/v2"
	"time"

	"github.com/labstack/echo/v4"
)

type ApiKind uint

const (
	REST ApiKind = iota
	HTMX
)

type ApiServer struct {
	E      *echo.Echo
	Kind   ApiKind
	Config ApiConfig
	Rand   *rand.PCG

	// groups     map[string]*echo.Group
	// middleware []echo.MiddlewareFunc
}

type ApiConfig struct {
	CORS []string `toml:"cors"`
	DB   DB

	TokenSecret          string        `toml:"token_secret"`
	TokenAccessValidity  time.Duration `toml:"token_access_validity"`
	TokenRefreshValidity time.Duration `toml:"token_refresh_validity"`

	LocalAuth         bool `toml:"local_auth"`
	OAuthEnabled      bool `tobl:"oauth_enabled"`
	AllowRegistration bool `toml:"allow_registration"`

	// Used for logout redirect and when no valid oauth callback referrer
	AppURI string `toml:"api_uri"`
}
