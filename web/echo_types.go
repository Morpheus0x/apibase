package web

import "github.com/labstack/echo/v4"

type ApiServer struct {
	e *echo.Echo

	// groups     map[string]*echo.Group
	// middleware []echo.MiddlewareFunc
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

// type HandleFunc func(c echo.Context) error
// type HandleFunc echo.HandlerFunc
