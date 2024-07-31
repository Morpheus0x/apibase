package web

import "github.com/labstack/echo/v4"

type ApiServer struct {
	e *echo.Echo

	// groups     map[string]*echo.Group
	// middleware []echo.MiddlewareFunc
}

//go:generate stringer -type Method -output ./stringer_method.go
type Method uint

const (
	GET Method = iota
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
