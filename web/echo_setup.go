package web

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.cc/apibase/log"
)

func SetupRest(config ApiConfig) *ApiServer {
	api := &ApiServer{e: echo.New(), kind: REST, config: config}
	if len(config.CORS) < 1 {
		api.config.CORS = []string{"*"}
	}
	api.e.HideBanner = true
	api.e.HidePort = true
	api.e.Use(middleware.Logger()) // TODO: replace with custom logger
	// TODO: replace all commented out e.Logger() calls with custom logger where useful
	api.e.Use(middleware.Recover())
	api.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: api.config.CORS,
	}))
	api.registerRestDefaultEndpoints()
	return api
}

func (api *ApiServer) registerRestDefaultEndpoints() log.Err {
	api.e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "No Auth Required!"})
	})

	api.e.POST("/auth/login", authLogin(api.config))
	api.e.GET("/auth/logout", authLogout(api.config), authJWT(api.config))
	api.e.POST("/auth/signup", authSignup(api.config))

	v1 := api.e.Group("/api/", authJWT(api.config))
	v1.GET("", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Welcome!"})
	})
	// Create default routes for login and general user flow
	return log.ErrorNil()
}

func (api *ApiServer) StartRest(bind string) log.Err {
	fmt.Printf("Rest API Server started on '%s'\n\n", bind) // TODO: replace with custom logger

	err := api.e.Start(bind) // blocking
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

func (api *ApiServer) Register(method HttpMethod, path string, handle echo.HandlerFunc) log.Err {
	if api.e == nil {
		return log.ErrorNew(log.ErrWebApiNotInit, "ApiServer not initialized")
	}
	switch method {
	case GET:
		api.e.GET(path, handle) // , api.middleware...
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
