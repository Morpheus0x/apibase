package web

import (
	"fmt"
	"net/http"

	"math/rand/v2"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
	t "gopkg.cc/apibase/types"
	"gopkg.cc/apibase/web_auth"
)

func SetupRest(config t.ApiConfig) (*t.ApiServer, error) {
	// TODO: overall better error logging
	if err := db.ValidateDB(config.DB); err != nil {
		return nil, err
	}
	if err := db.MigrateDefaultTables(config.DB); err != nil {
		return nil, err
	}
	api := &t.ApiServer{
		E:      echo.New(),
		Kind:   t.REST,
		Config: config,
		Rand:   rand.NewPCG(rand.Uint64(), rand.Uint64()),
	}
	if len(config.CORS) < 1 {
		api.Config.CORS = []string{"*"} // TODO: maybe error instead of assuming *
	}
	api.E.HideBanner = true
	api.E.HidePort = true
	api.E.Use(middleware.Logger()) // TODO: replace with custom logger
	// TODO: replace all commented out e.Logger() calls with custom logger where useful
	api.E.Use(middleware.Recover())
	api.E.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: api.Config.CORS,
	}))
	RegisterRestDefaultEndpoints(api)
	return api, nil
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

func StartRest(api *t.ApiServer, bind string) log.Err {
	fmt.Printf("Rest API Server started on '%s'\n\n", bind) // TODO: replace with custom logger

	err := api.E.Start(bind) // blocking
	if err != nil {
		return log.ErrorNew(log.ErrWebBind, "unable to start rest api with bind '%s': %v", bind, err)
	}
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

func Register(api *t.ApiServer, method t.HttpMethod, path string, handle echo.HandlerFunc) log.Err {
	if api.E == nil {
		return log.ErrorNew(log.ErrWebApiNotInit, "ApiServer not initialized")
	}
	switch method {
	case t.GET:
		api.E.GET(path, handle) // , api.middleware...
	default:
		return log.ErrorNew(log.ErrWebUnknownMethod, "Unknown Method")
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
