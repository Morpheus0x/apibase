package web

import (
	"embed"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/errx"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	wr "gopkg.cc/apibase/web_response"
)

type CustomURI struct {
	uri *url.URL
}

// Clones the specified *url.URL struct into *CustomURI
func NewCustomUri(uri *url.URL) *CustomURI {
	return &CustomURI{uri: h.CloneURL(uri)}
}

func NewCustomUriFromString(uri string) (*CustomURI, error) {
	parsedUri, err := url.Parse(uri)
	if err != nil {
		return &CustomURI{}, errx.Wrap(err, "Unable to create new CustomURI")
	}
	return &CustomURI{uri: parsedUri}, nil
}

func (app CustomURI) String() string {
	return app.uri.String()
}

func (app *CustomURI) JoinPath(path string) *CustomURI {
	app.uri.Path = app.uri.JoinPath(path).Path
	return app
}

// Supports string, int and web_response.ResponseId as value types
func (app *CustomURI) AddQueryParam(key string, value any) *CustomURI {
	query := app.uri.Query()
	if query.Has(key) {
		query.Del(key)
	}
	query.Add(key, parseQueryParam(value))
	app.uri.RawQuery = query.Encode()
	return app
}

// Supports string, int and web_response.ResponseId as value types
func (app *CustomURI) AddQueryParams(params []QueryParam[any]) *CustomURI {
	query := app.uri.Query()
	for _, p := range params {
		if query.Has(p.Key) {
			query.Del(p.Key)
		}
		query.Add(p.Key, p.Value())
	}
	app.uri.RawQuery = query.Encode()
	return app
}

// Should return pointer to struct containing custom access claim data.
// The struct must have json tags and you are encouraged to obfuscate the tags e.g. using a, b, c, ...
// If bool variable is true, return only an initialized empty struct of the desired type, this must always return without an error.
// Errors will be logged and access claim data will be nil in jwt.
type AccessClaimDataFunc func(int, bool) (any, error)

type ApiServer struct {
	E      *echo.Echo  // Direct access to the echo webserver instance
	Api    *echo.Group // Used to register API endpoints, leading slash already present (/api/<route>)
	Kind   ApiKind     // REST (or HTMX, TODO: this)
	Config ApiConfig   // API config used to initialize ApiServer
	DB     db.DB       // Database connection for this ApiServer

	accessClaimData AccessClaimDataFunc // Custom Access Claims for User
	// middleware []echo.MiddlewareFunc
}

// Use this to add your own access token claim data, see web.AccessClaimDataFunc for further details.
// To get additinal required values to determine access claim data for user, use dependency injection on web.AccessClaimDataFunc
func (api *ApiServer) RegisterAccessClaimDataFunc(accessClaimDataFunc AccessClaimDataFunc) {
	api.accessClaimData = accessClaimDataFunc
}

func (api ApiServer) GetAccessClaimData(userId int) any {
	if api.accessClaimData == nil {
		return nil
	}
	data, err := api.accessClaimData(userId, false)
	if err != nil {
		log.Logf(log.LevelError, "Unable to get custom access claim data for user (id: %d): %s", userId, err.Error())
		return nil
	}
	return data
}

// This is required for parsing access token, since it requires memory to be already allocated for the claims
func (api ApiServer) GetAccessClaimDataType() any {
	if api.accessClaimData == nil {
		return nil
	}
	data, _ := api.accessClaimData(0, true)
	return data
}

type ApiConfig struct {
	CORS    []string `toml:"cors"`
	ApiBind string   `toml:"api_bind"`
	AppURI  string   `toml:"app_uri"` // Used for logout redirect and when no valid oauth callback referrer

	// Secrets
	TokenSecret h.SecretString `toml:"token_secret"`

	// Flags
	LocalAuth          bool `toml:"local_auth"`
	OAuthEnabled       bool `toml:"oauth_enabled"`
	AllowRegistration  bool `toml:"allow_registration"`
	ReqireConfirmEmail bool `toml:"require_confirmed_email"` // TODO: this // Before user is allowed to login

	// Nested Structs
	ApiRoot  RootOptions        `toml:"api_root"` // Configure the apibase root behaviour (local, static, (reverse) proxy, or embedfs)
	Settings *ApiConfigSettings `toml:"settings"`

	// Internal Data
	tokenSecretBytes []byte     // decoded from TokenSecret string
	appURI           *CustomURI // will be parsed from ApiConfig.AppURI
	embedFS          embed.FS   // if configured in ApiRoot, must be registered with ApiConfig.RegisterEmbedFS()
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

// Get *url.URL from AppURI, panics if ApiConfig.AppURI couldn't be parsed
func (ac *ApiConfig) AppUri() *CustomURI {
	if ac.appURI != nil && ac.appURI.uri != nil {
		return NewCustomUri(ac.appURI.uri)
	}

	uri, err := url.ParseRequestURI(ac.AppURI)
	if err != nil {
		log.Logf(log.LevelCritical, "app_uri (from config: %s) must be valid url with protocol and without fragment of the application using the api: %s", ac.AppURI, err.Error())
		panic(1)
	}
	ac.appURI = NewCustomUri(uri)

	return NewCustomUri(ac.appURI.uri)
}

func (ac *ApiConfig) RegisterEmbedFS(embedFS embed.FS) error {
	if ac.ApiRoot.Kind != FsEmbed {
		return errx.NewWithType(ErrFsKindNotEmbed, "unable to register EmbedFS")
	}
	ac.embedFS = embedFS
	return nil
}

// used to protect internal embedFS ApiConfig member, only use this if ApiRoot Kind embedfs is configured
func (ac *ApiConfig) GetEmbedFS() embed.FS {
	return ac.embedFS
}

func (ac ApiConfig) AddCookieExpiryMargin(validity time.Duration) time.Duration {
	return validity + time.Duration(float32(validity)*ac.Settings.TokenCookieExpiryMargin)
}

type ApiConfigSettings struct {
	TomlTokenAccessValidity          string `toml:"token_access_validity"`
	TomlTokenRefreshValidity         string `toml:"token_refresh_validity"`
	TomlTokenCookieExpiryMargin      string `toml:"token_cookie_expiry_margin"`
	TomlTokenAccessRenewMargin       string `toml:"token_access_renew_margin"`
	TomlTokenRefreshRenewMargin      string `toml:"token_refresh_renew_margin"`
	TomlTimeoutSubprocStartup        string `toml:"timeout_subproc_startup"`
	TomlTimeoutSubprocShutdown       string `toml:"timeout_subproc_shutdown"`
	TomlTimeoutScheduledTaskStartup  string `toml:"timeout_scheduled_task_startup"`
	TomlTimeoutScheduledTaskShutdown string `toml:"timeout_scheduled_task_shutdown"`

	TokenAccessValidity          time.Duration `internal:"token_access_validity"`
	TokenRefreshValidity         time.Duration `internal:"token_refresh_validity"`
	TokenCookieExpiryMargin      float32       `internal:"token_cookie_expiry_margin" parsetype:"percentage"`
	TokenAccessRenewMargin       time.Duration `internal:"token_access_renew_margin"`
	TokenRefreshRenewMargin      time.Duration `internal:"token_refresh_renew_margin"`
	TimeoutSubprocStartup        time.Duration `internal:"timeout_subproc_startup"`
	TimeoutSubprocShutdown       time.Duration `internal:"timeout_subproc_shutdown"`
	TimeoutScheduledTaskStartup  time.Duration `internal:"timeout_scheduled_task_startup"`
	TimeoutScheduledTaskShutdown time.Duration `internal:"timeout_scheduled_task_shutdown"`
}

func (settings *ApiConfigSettings) AddMissingFromDefaults() error {
	defaults := &ApiConfigSettings{
		TokenAccessValidity:          time.Minute * 15,
		TokenRefreshValidity:         time.Hour * 24 * 30,
		TokenCookieExpiryMargin:      0.2, // 20%
		TokenAccessRenewMargin:       time.Minute,
		TokenRefreshRenewMargin:      time.Hour * 24 * 7,
		TimeoutSubprocStartup:        time.Second,
		TimeoutSubprocShutdown:       time.Second * 3,
		TimeoutScheduledTaskStartup:  time.Second,
		TimeoutScheduledTaskShutdown: time.Second * 60,
	}
	return h.ParseTomlConfigAndDefaults(settings, defaults)
}

//go:generate stringer -type HttpMethod -output ./stringer_HttpMethod.go
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

type FileSystemKind string

const (
	FsLocal  FileSystemKind = "local"
	FsStatic FileSystemKind = "static"
	FsProxy  FileSystemKind = "proxy"
	FsEmbed  FileSystemKind = "embedfs"
)

type RootOptions struct {
	Kind   FileSystemKind `toml:"kind"`
	Target string         `toml:"target"` // If Kind is FsEmbed, Target must start with the embedded folder name (without leading ./)
}

func (ac *ApiConfig) ValidateApiRoot() error {
	switch ac.ApiRoot.Kind {
	case FsLocal:
		return nil
	case FsStatic:
		path, err := filepath.Abs(ac.ApiRoot.Target)
		if err != nil {
			return errx.Wrapf(err, "unable to validate RootOptions static")
		}
		if _, err := os.Stat(path); err != nil {
			return errx.Wrap(err, "unable to validate RootOptions static")
		}
		return nil
	case FsProxy:
		url, err := url.Parse(ac.ApiRoot.Target)
		if err != nil {
			return errx.Wrap(err, "unable to validate RootOptions proxy")
		}
		if url.Scheme != "http" && url.Scheme != "https" {
			return errx.New("schema for RootOptions proxy must be http(s)")
		}
		// TODO: maybe add additional validation for url
		return nil
	case FsEmbed:
		rootDir, err := ac.embedFS.ReadDir(ac.ApiRoot.Target)
		if err != nil {
			return errx.Wrap(err, "unable to validate RootOptions embedfs")
		}
		if len(rootDir) < 1 {
			return errx.Newf("unable to validate RootOptions embedfs: embedded folder '%s' is empty", ac.ApiRoot.Target)
		}
		return nil
	default:
		return errx.New("no RootOptions Kind specified, must be local, static, proxy or embedfs")
	}
}

// used for correct forwarding after oauth callback
type StateReferrer struct {
	Nonce string `json:"nonce"`
	URI   string `json:"uri"`
}

// Supports string, int and web_response.ResponseId as UntypedValue types
type QueryParam[T any] struct {
	Key          string
	UntypedValue T
}

func (p QueryParam[T]) Value() string {
	return parseQueryParam(p.UntypedValue)
}

// Supports string, int and web_response.ResponseId as value types
func parseQueryParam(value any) string {
	switch v := any(value).(type) {
	case string:
		return v
	case int, wr.ResponseId:
		return fmt.Sprintf("%d", v)
	default:
		return ""
	}
}
