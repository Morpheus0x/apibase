package web

import (
	"context"
	"net/http"
	"time"

	"math/rand/v2"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/web_auth"
	"gopkg.cc/apibase/webtype"
	t "gopkg.cc/apibase/webtype"
)

func SetupRest(config t.ApiConfig) (*t.ApiServer, *log.Error) {
	if err := db.ValidateDB(config.DB); !err.IsNil() {
		return nil, err.Extend("unable to setup rest api")
	}
	if err := db.MigrateDefaultTables(config.DB); !err.IsNil() {
		return nil, err.Extend("unable to migrate db tables")
	}
	if len(config.TokenSecret) < 32 {
		return nil, log.NewError("TokenSecret must be at least 32 characters")
	}
	if !config.LocalAuth && !config.OAuthEnabled {
		return nil, log.NewError("No Authentication method enabled, either LocalAuth, OAuthEnabled or both need to be enabled")
	}
	if config.AppURI == "" {
		return nil, log.NewError("No AppURI specified, this must be a fully qualified uri of the application using the api")
	}
	api := &t.ApiServer{
		E:      echo.New(),
		Kind:   t.REST,
		Config: config,
		Rand:   rand.NewPCG(rand.Uint64(), rand.Uint64()),
	}
	if len(config.CORS) < 1 {
		log.Log(log.LevelWarning, "CORS is not set, assuming '*', this should not be used in a production environment!")
		api.Config.CORS = []string{"*"}
	}
	if config.TokenAccessValidity == "" {
		log.Logf(log.LevelWarning, "AccessTokenValidity is not set, assuming default validity of %s", webtype.TOKEN_ACCESS_VALIDITY.String())
		api.Config.TokenAccessValidity = webtype.TOKEN_ACCESS_VALIDITY.String()
	}
	if config.TokenRefreshValidity == "" {
		log.Logf(log.LevelWarning, "TokenRefreshValidity is not set, assuming default validity of %s", webtype.TOKEN_REFRESH_VALIDITY.String())
		api.Config.TokenRefreshValidity = webtype.TOKEN_REFRESH_VALIDITY.String()
	}
	api.E.HideBanner = true
	api.E.HidePort = true
	api.E.Use(log.EchoLoggerMiddleware(log.LevelDebug))
	api.E.Use(middleware.Recover())
	api.E.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: api.Config.CORS,
	}))
	RegisterRestDefaultEndpoints(api)
	return api, log.ErrorNil()
}

func RegisterRestDefaultEndpoints(api *t.ApiServer) {
	api.E.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "No Auth Required!"})
	})

	v1 := api.E.Group("/api/", web_auth.AuthJWT(api))
	v1.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Welcome!"})
	})
}

// start rest api server, is blocking
func StartRestBlocking(api *t.ApiServer, bind string) *log.Error {
	log.Logf(log.LevelNotice, "Rest API Server starting on '%s'", bind)

	err := api.E.Start(bind) // blocking
	if err != nil {
		return log.NewErrorWithTypef(ErrWebBind, "'%s': %v", bind, err)
	}
	return log.ErrorNil()
}

// start rest api server, is non-blocking
func StartRest(api *t.ApiServer, bind string, shutdown chan struct{}, next chan struct{}) *log.Error {
	go func() {
		<-shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // TODO: remove hardcoded timeout
		defer cancel()
		err := api.E.Shutdown(ctx)
		if err != nil {
			log.NewErrorWithType(ErrWebShutdown, err.Error()).Log()
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
	startErrorChan := make(chan *log.Error)

	go func() {
		err := api.E.Start(bind) // blocking
		if err != nil && err != http.ErrServerClosed {
			// Possible, but very unlikely race-condition with checking a channel this way
			select {
			case startErrorChan <- log.NewErrorWithTypef(ErrWebBind, "'%s': %v", bind, err):
			default:
			}
		}
	}()

	select {
	case err := <-startErrorChan:
		if err != nil {
			return err
		}
	case <-ctx.Done():
	}
	// for race-condition, close must happen exactly after echo start error catching select checks the channel but before it writes to it, very unlikely
	close(startErrorChan)
	log.Logf(log.LevelNotice, "Rest API Server started on '%s'", bind)
	return log.ErrorNil()
}

// func SetupStatic(root string) *ApiServer {
// 	if root == "" { // TODO: validate root is path to folder
// 		log.Panic(log.ErrEmptyString, "root '%s' must be valid path to folder", root)
// 	}
// 	api := &ApiServer{e: echo.New()}
// 	api.e.Static("/", root)
// 	return api
// }

func Register(api *t.ApiServer, method t.HttpMethod, path string, handle echo.HandlerFunc) *log.Error {
	if api.E == nil {
		return log.NewErrorWithType(ErrWebApiNotInit, "")
	}
	switch method {
	case t.GET:
		api.E.GET(path, handle) // , api.middleware...
	default:
		return log.NewErrorWithTypef(ErrWebUnknownMethod, "types.HttpMethod(%d)", method)
	}
	return log.ErrorNil()
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
