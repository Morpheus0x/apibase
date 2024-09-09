package web

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/app"
)

type ApiServer struct {
	e      *echo.Echo
	kind   ApiKind
	config app.ApiConfig

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

type ApiKind uint

const (
	REST ApiKind = iota
	HTMX
)

type UserRole uint

const (
	OrgViewer UserRole = iota
	OrgUser
	OrgAdmin
	SuperAdmin = 99
)

type jwtAccessClaims struct {
	Name       string   `json:"name"`
	Role       UserRole `json:"role"`
	CSRFHeader string   `json:"csrf_header"`
	jwt.RegisteredClaims
}

type jwtRefreshClaims struct {
	Name       string   `json:"name"`
	Role       UserRole `json:"role"`
	CSRFHeader string   `json:"csrf_header"`
	jwt.RegisteredClaims
}

// type HandleFunc func(c echo.Context) error
// type HandleFunc echo.HandlerFunc
