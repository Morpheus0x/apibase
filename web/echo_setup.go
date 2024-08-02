package web

import (
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/log"
)

func Setup() *ApiServer {
	api := &ApiServer{e: echo.New()}

	return api
}

// func SetupStatic(root string) *ApiServer {
// 	if root == "" { // TODO: validate root is path to folder
// 		log.Panic(log.ErrEmptyString, "root '%s' must be valid path to folder", root)
// 	}
// 	api := &ApiServer{e: echo.New()}
// 	api.e.Static("/", root)
// 	return api
// }

func (api *ApiServer) Register(method Method, path string, handle echo.HandlerFunc) log.Err {
	if api.e == nil {
		return log.ErrorNew(log.ErrWebApiNotInit, "ApiServer not initialized")
	}
	switch method {
	case GET:
		api.e.GET(path, handle) // , api.middleware...
	default:
		return log.ErrorNew(log.ErrWebUnknownMethod, "Unknown Method")
	}

	return api.registerDefaultEndpoints()
}

func (api *ApiServer) registerDefaultEndpoints() log.Err {
	// Create default routes for login and general user flow
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
