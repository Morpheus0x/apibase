package web

import (
	"encoding/base64"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/xhit/go-str2duration/v2"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/errx"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
)

type ApiServer struct {
	E      *echo.Echo  // Direct access to the echo webserver instance
	Api    *echo.Group // Used to register API endpoints, leading slash already present (/api/<route>)
	Kind   ApiKind     // REST (or HTMX, TODO: this)
	Config ApiConfig   // API config used to initialize ApiServer
	DB     db.DB       // Database connection for this ApiServer

	// middleware []echo.MiddlewareFunc
}

type ApiConfig struct {
	CORS []string `toml:"cors"`

	TokenSecret          h.SecretString `toml:"token_secret"`
	TokenAccessValidity  string         `toml:"token_access_validity"`
	TokenRefreshValidity string         `toml:"token_refresh_validity"`

	LocalAuth          bool `toml:"local_auth"`
	OAuthEnabled       bool `toml:"oauth_enabled"`
	AllowRegistration  bool `toml:"allow_registration"`
	ReqireConfirmEmail bool `toml:"require_confirmed_email"` // TODO: this // Before user is allowed to login

	DefaultOrgID int `toml:"default_org_id"`

	ApiBind string `toml:"api_bind"`
	AppURI  string `toml:"app_uri"` // Used for logout redirect and when no valid oauth callback referrer

	// Nested Structs
	ApiRoot        RootOptions        `toml:"api_root"` // Configure the apibase root behaviour (local, static, (reverse) proxy)
	DefaultOrgRole map[string]JwtRole `toml:"default_org_role"`

	// Internal Data
	tokenSecretBytes []byte // decoded from TokenSecret string
}

func (ac ApiConfig) TokenSecretBytes() []byte {
	if len(ac.tokenSecretBytes) > 0 {
		return ac.tokenSecretBytes
	}
	secret, err := base64.StdEncoding.DecodeString(ac.TokenSecret.GetSecret())
	if err != nil {
		// is allowed to panic, since this should't occur if TokenSecret string is parsed during ApiConfig setup
		log.Logf(log.LevelCritical, "token secret isn't a base64 string: %s", err.Error())
		// TODO: also print stack trace
		panic(1)
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
	duration, err := str2duration.ParseDuration(ac.TokenRefreshValidity)
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

type RootOptions struct {
	Kind   string `toml:"kind"`
	Target string `toml:"target"`
}

func (ro RootOptions) Validate() error {
	switch ro.Kind {
	case "local":
		return nil
	case "static":
		path, err := filepath.Abs(ro.Target)
		if err != nil {
			return errx.Wrapf(err, "unable to validate RootOptions static")
		}
		if _, err := os.Stat(path); err != nil {
			return errx.Wrap(err, "unable to validate RootOptions static")
		}
		return nil
	case "proxy":
		url, err := url.Parse(ro.Target)
		if err != nil {
			return errx.Wrap(err, "unable to validate RootOptions proxy")
		}
		if url.Scheme != "http" && url.Scheme != "https" {
			return errx.New("schema for RootOptions proxy must be http(s)")
		}
		// TODO: maybe add additional validation for url
		return nil
	default:
		return errx.New("no RootOptions Kind specified, must be local, static or proxy")
	}
}

// used for correct forwarding after oauth callback
type StateReferrer struct {
	Nonce string `json:"nonce"`
	URI   string `json:"uri"`
}
