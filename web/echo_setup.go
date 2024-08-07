package web

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.cc/apibase/log"
)

func SetupRest(config ApiConfig) *ApiServer {
	api := &ApiServer{e: echo.New(), kind: REST, config: config}
	if len(config.CORS) < 1 {
		api.config.CORS = []string{"*"}
	}
	api.e.Use(middleware.Logger())
	api.e.Use(middleware.Recover())
	api.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: api.config.CORS,
	}))
	api.registerRestDefaultEndpoints()
	return api
}

func (api *ApiServer) registerRestDefaultEndpoints() log.Err {
	// Create middleware requireing JWT authorization
	api.e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "No Auth Required!")
	})
	api.e.POST("/api/v1/auth/login", login)
	// api.e.POST("/api/v1/auth/signup", defaultEndpointSignup)
	v1 := api.e.Group("/api/v1/", authMiddlewareWrapper, echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtAccessClaims)
		},
		TokenLookup: "cookie:access_token", // "header:Authorization:Bearer ,cookie:access_token",
		SigningKey:  []byte("superSecretSecret"),
	}))
	v1.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome")
	})
	// Create default routes for login and general user flow
	return log.ErrorNil()
}

func (api *ApiServer) StartRest(bind string) log.Err {
	err := api.e.Start(bind)
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
