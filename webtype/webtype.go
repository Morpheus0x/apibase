package webtype

import (
	"github.com/golang-jwt/jwt/v5"
)

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

type UserRole uint

const (
	OrgViewer UserRole = iota
	OrgUser
	OrgAdmin
	SuperAdmin = 99
)

type JwtAccessClaims struct {
	Name       string   `json:"name"`
	Role       UserRole `json:"role"`
	CSRFHeader string   `json:"csrf_header"`
	jwt.RegisteredClaims
}

type JwtRefreshClaims struct {
	Name       string   `json:"name"`
	Role       UserRole `json:"role"`
	CSRFHeader string   `json:"csrf_header"`
	jwt.RegisteredClaims
}

// type HandleFunc func(c echo.Context) error
// type HandleFunc echo.HandlerFunc

type StateReferrer struct {
	Nonce string `json:"nonce"`
	URI   string `json:"uri"`
}
