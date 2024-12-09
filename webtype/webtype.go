package webtype

import (
	"math/rand/v2"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/db"
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
	DB   db.DB

	TokenSecret          string        `toml:"token_secret"`
	TokenAccessValidity  time.Duration `toml:"token_access_validity"`
	TokenRefreshValidity time.Duration `toml:"token_refresh_validity"`

	LocalAuth         bool `toml:"local_auth"`
	OAuthEnabled      bool `tobl:"oauth_enabled"`
	AllowRegistration bool `toml:"allow_registration"`

	// Used for logout redirect and when no valid oauth callback referrer
	AppURI string `toml:"api_uri"`
}

//go:generate stringer -type HttpMethod -output ./stringer_httpmethod.go
type HttpMethod uint

const (
	GET HttpMethod = iota
	HEAD
	POST
	PUT
	DELETE
	CONNECT
	OPTIONS
	TRACE
	PATCH
)

type ApiKind uint

const (
	REST ApiKind = iota
	HTMX
)

type UserRole uint

const (
	OrgViewer UserRole = iota
	OrgUser
	OrgAdmin
	SuperAdmin = 99
)

type JwtAccessClaims struct {
	Name       string   `json:"name"`
	Role       UserRole `json:"role"`
	CSRFHeader string   `json:"csrf_header"`
	jwt.RegisteredClaims
}

type JwtRefreshClaims struct {
	Name       string   `json:"name"`
	Role       UserRole `json:"role"`
	CSRFHeader string   `json:"csrf_header"`
	jwt.RegisteredClaims
}

// type HandleFunc func(c echo.Context) error
// type HandleFunc echo.HandlerFunc

type StateReferrer struct {
	Nonce string `json:"nonce"`
	URI   string `json:"uri"`
}
