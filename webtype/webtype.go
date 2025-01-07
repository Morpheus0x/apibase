package webtype

import (
	"encoding/base64"
	"math/rand/v2"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xhit/go-str2duration/v2"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
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

	TokenSecret          string `toml:"token_secret"`
	TokenAccessValidity  string `toml:"token_access_validity"`
	TokenRefreshValidity string `toml:"token_refresh_validity"`

	LocalAuth          bool `toml:"local_auth"`
	OAuthEnabled       bool `tobl:"oauth_enabled"`
	AllowRegistration  bool `toml:"allow_registration"`
	ReqireConfirmEmail bool `toml:"require_confirmed_email"` // TODO: this // Before user is allowed to login
	DefaultOrgID       int  `toml:"default_org_id"`

	ApiBind string `toml:"api_bind"`
	AppURI  string `toml:"app_uri"` // Used for logout redirect and when no valid oauth callback referrer

	tokenSecretBytes []byte // decoded from TokenSecret string
}

func (ac ApiConfig) TokenSecretBytes() []byte {
	if len(ac.tokenSecretBytes) > 0 {
		return ac.tokenSecretBytes
	}
	secret, err := base64.StdEncoding.DecodeString(ac.TokenSecret)
	if err != nil {
		log.Logf(log.LevelCritical, "token secret isn't a base64 string: %s", err.Error())
		panic(1) // is allowed to panic, since this souldn't occur if TokenSecret string is parsed during ApiConfig setup
	}
	ac.tokenSecretBytes = secret
	return ac.tokenSecretBytes
}

func (ac ApiConfig) TokenAccessValidityDuration() time.Duration {
	duration, err := str2duration.ParseDuration(ac.TokenAccessValidity)
	if err != nil {
		log.Logf(log.LevelCritical, "unable to parse TokenAccessValidity duration: %s, assuming default %s", ac.TokenAccessValidity, TOKEN_ACCESS_VALIDITY.String())
		return TOKEN_ACCESS_VALIDITY
	}
	return duration
}

func (ac ApiConfig) TokenRefreshValidityDuration() time.Duration {
	duration, err := str2duration.ParseDuration(ac.TokenAccessValidity)
	if err != nil {
		log.Logf(log.LevelCritical, "unable to parse TokenRefreshValidity duration: %s, assuming default %s", ac.TokenRefreshValidity, TOKEN_REFRESH_VALIDITY.String())
		return TOKEN_REFRESH_VALIDITY
	}
	return duration
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

// type HandleFunc func(c echo.Context) error
// type HandleFunc echo.HandlerFunc

type StateReferrer struct {
	Nonce string `json:"nonce"`
	URI   string `json:"uri"`
}
