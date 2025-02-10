package web_setup

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/web"
	"gopkg.cc/apibase/web_auth"
	"gopkg.cc/apibase/web_oauth"
)

func SetupRest(config web.ApiConfig, database db.DB) (*web.ApiServer, error) {
	if err := db.ValidateDB(database); err != nil {
		return nil, errx.Wrap(err, "unable to setup rest api")
	}
	if err := db.MigrateDefaultTables(database); err != nil {
		return nil, errx.Wrap(err, "unable to migrate db tables")
	}
	tokenSecretBytes, err := base64.StdEncoding.DecodeString(config.TokenSecret.GetSecret())
	if err != nil || len(tokenSecretBytes) < 64 {
		return nil, errx.New("TokenSecret must be random base64 byte string with at least 64 bytes")
	}
	if !config.LocalAuth && !config.OAuthEnabled {
		return nil, errx.New("No Authentication method enabled, either LocalAuth, OAuthEnabled or both need to be enabled")
	}
	if _, err := url.ParseRequestURI(config.AppURI); err != nil {
		return nil, errx.Newf("AppURI (from config: %s) must be valid uri with protocol and without fragment of the application using the api: %s", config.AppURI, err.Error())
	}
	if err := config.ApiRoot.Validate(); err != nil {
		return nil, err
	}
	api := &web.ApiServer{
		E: echo.New(),
		// Api: will be set by RegisterRestDefaultEndpoints()
		Kind:   web.REST,
		Config: config,
		DB:     database,
	}
	if len(config.CORS) < 1 {
		log.Log(log.LevelWarning, "CORS is not set, assuming '*', this should not be used in a production environment!")
		api.Config.CORS = []string{"*"}
	}
	if config.TokenAccessValidity == "" {
		log.Logf(log.LevelWarning, "AccessTokenValidity is not set, assuming default validity of %s", web.TOKEN_ACCESS_VALIDITY.String())
		api.Config.TokenAccessValidity = web.TOKEN_ACCESS_VALIDITY.String()
	}
	if config.TokenRefreshValidity == "" {
		log.Logf(log.LevelWarning, "TokenRefreshValidity is not set, assuming default validity of %s", web.TOKEN_REFRESH_VALIDITY.String())
		api.Config.TokenRefreshValidity = web.TOKEN_REFRESH_VALIDITY.String()
	}
	api.E.HideBanner = true
	api.E.HidePort = true
	api.E.Use(log.EchoLoggerMiddleware(log.LevelDebug, log.LevelDevel))
	api.E.HTTPErrorHandler = web.EchoErrorHandler
	api.E.Use(middleware.Recover())
	api.E.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: api.Config.CORS,
	}))
	RegisterRestDefaultEndpoints(api)
	if api.Config.LocalAuth {
		web_auth.RegisterAuthEndpoints(api)
	}
	if api.Config.OAuthEnabled {
		web_oauth.RegisterOAuthEndpoints(api)
	}
	return api, nil
}

func RegisterRestDefaultEndpoints(api *web.ApiServer) {
	switch api.Config.ApiRoot.Kind {
	case "local":
		api.E.GET("/", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"message": "apibase"})
		})
		// TODO: change default local response and optionally add additional routes
	case "static":
		api.E.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:   api.Config.ApiRoot.Target,
			Index:  "index.html",
			Browse: false,
		}))
	case "proxy":
		url, err := url.Parse(api.Config.ApiRoot.Target)
		if err != nil {
			// Allowed to panic
			log.Logf(log.LevelCritical, "ApiRoot proxy must contain valid uri target: %s", err.Error())
			panic(1)
		}
		api.E.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
			Skipper: func(e echo.Context) bool {
				path := e.Request().URL.Path
				return strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/auth")
			},
			Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{{
				URL: url,
			}}),
			ErrorHandler: func(c echo.Context, err error) error {
				echoErr := &echo.HTTPError{}
				valid := errors.As(err, &echoErr)
				if valid && echoErr.Code > 499 && echoErr.Code < 600 {
					return c.HTML(http.StatusBadGateway, web.GatewayTimeoutHTML)
				}
				return err
			},
		}))
	default:
		// Allowed to panic
		errx.New("ApiRoot.Kind must be local, static or proxy. This should have been verified during setup!")
		panic(1)
	}

	api.E.GET("/auth/csrf_token", web.GetCSRF(api))
	apiGroup := api.E.Group("/api/", web.CheckCSRF(api), web.AuthJWT(api))
	apiGroup.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Welcome!"})
	})
	api.Api = apiGroup
}

// start rest api server, is blocking
func StartRestBlocking(api *web.ApiServer, bind string) error {
	log.Logf(log.LevelNotice, "Rest API Server starting on '%s'", bind)

	err := api.E.Start(bind) // blocking
	if err != nil {
		return errx.NewWithTypef(ErrWebBind, "'%s': %v", bind, err)
	}
	return nil
}

// start rest api server, is non-blocking, require shutdown and next channels for clean shutdown, these can be created using base.ApiBase[T].GetCloseStageChannels()
func StartRest(api *web.ApiServer, bind string, shutdown chan struct{}, next chan struct{}) error {
	abort := make(chan struct{})

	go func() { // echo shutdown
		select {
		case <-shutdown:
		case <-abort:
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // TODO: remove hardcoded timeout
		defer cancel()
		err := api.E.Shutdown(ctx)
		if err != nil {
			log.Logf(log.LevelError, "%s", errx.WrapWithType(ErrWebShutdown, err, "").Error())
		} else {
			log.Log(log.LevelNotice, "Rest API Server shutdown successful.")
		}
		select {
		case <-next:
			log.Log(log.LevelError, "rest api next channel already closed, make sure channel staging is done correctly")
		default:
			close(next)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond) // TODO: remove hardcoded timeout
	defer cancel()
	startupError := struct {
		Chan       chan error
		sync.Mutex // to protect agains a VERY edge case race-condition
	}{Chan: make(chan error)}

	go func() { // echo start
		err := api.E.Start(bind) // blocking
		if err != nil && err != http.ErrServerClosed {
			startupError.Lock()
			select {
			case startupError.Chan <- errx.NewWithTypef(ErrWebBind, "'%s': %v", bind, err):
			default:
			}
			startupError.Unlock()
		}
	}()

	select {
	case err := <-startupError.Chan:
		if err != nil {
			close(abort)
			return err
		}
	case <-ctx.Done():
	}
	startupError.Lock()
	close(startupError.Chan)
	startupError.Unlock()

	log.Logf(log.LevelNotice, "Rest API Server started on '%s'", bind)
	return nil
}

// func SetupStatic(root string) *ApiServer {
// 	if root == "" { // TODO: validate root is path to folder
// 		log.Panic(log.ErrEmptyString, "root '%s' must be valid path to folder", root)
// 	}
// 	api := &ApiServer{e: echo.New()}
// 	api.e.Static("/", root)
// 	return api
// }

func Register(api *web.ApiServer, method web.HttpMethod, path string, handle echo.HandlerFunc) error {
	if api.E == nil {
		return errx.NewWithType(ErrWebApiNotInit, "")
	}
	switch method {
	case web.GET:
		api.E.GET(path, handle) // , api.middleware...
	default:
		return errx.NewWithTypef(ErrWebUnknownMethod, "types.HttpMethod(%d)", method)
	}
	return nil
}

// // Serve static folder or files in destination at provided path
// func (api *ApiServer) RegisterStatic(path string, destination string) log.Err {
// 	if path == "" || destination == "" {
// 		return log.ErrorNew(log.ErrEmptyString, "path '%s' or destination '%s' is empty", path, destination)
// 	}
// 	api.e.Static(path, destination)
// 	return log.ErrorNil()
// }

// func (api *ApiServer) RegisterGroup(name string, path string) log.Err {
// 	if _, exists := api.groups[name]; exists {
// 		return log.ErrorNew(log.ErrWebGroupExists, "group for name '%s' already exists", name)
// 	}
// 	api.groups[name] = api.e.Group(path)
// 	return log.ErrorNil()
// }

// func (api *ApiServer) RegisterGroupMiddleware(name string, fn ...echo.MiddlewareFunc) log.Err {
// 	if _, ok := api.groups[name]; !ok {
// 		return log.ErrorNew(log.ErrWebGroupNotExists, "group for name '%s' doesn't exist", name)
// 	}
// 	if len(fn) < 1 {
// 		return log.ErrorNew(log.ErrEmptyArray, "fn must contain at least on middleware function")
// 	}
// 	api.groups[name].Use(fn...)
// 	return log.ErrorNil()
// }
