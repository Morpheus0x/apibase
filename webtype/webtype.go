package webtype

import (
	"encoding/base64"
	"math/rand/v2"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/xhit/go-str2duration/v2"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/tables"
)

type ApiServer struct {
	E      *echo.Echo
	Kind   ApiKind
	Config ApiConfig
	Rand   *rand.PCG

	// groups     map[string]*echo.Group
	// middleware []echo.MiddlewareFunc
}

type ApiConfig struct {
	CORS []string `toml:"cors"`
	DB   db.DB

	TokenSecret          string `toml:"token_secret"`
	TokenAccessValidity  string `toml:"token_access_validity"`
	TokenRefreshValidity string `toml:"token_refresh_validity"`

	LocalAuth         bool `toml:"local_auth"`
	OAuthEnabled      bool `tobl:"oauth_enabled"`
	AllowRegistration bool `toml:"allow_registration"`
	DefaultOrgID      int  `toml:"default_org_id"`

	ApiBind string `toml:"api_bind"`
	AppURI  string `toml:"app_uri"` // Used for logout redirect and when no valid oauth callback referrer

	tokenSecretBytes []byte // decoded from TokenSecret string
}

func (ac ApiConfig) TokenSecretBytes() []byte {
	if len(ac.tokenSecretBytes) > 0 {
		return ac.tokenSecretBytes
	}
	secret, err := base64.StdEncoding.DecodeString(ac.TokenSecret)
	if err != nil {
		log.Logf(log.LevelCritical, "token secret isn't a base64 string: %s", err.Error())
		panic(1) // is allowed to panic, since this souldn't occur if TokenSecret string is parsed during ApiConfig setup
	}
	ac.tokenSecretBytes = secret
	return ac.tokenSecretBytes
}

func (ac ApiConfig) TokenAccessValidityDuration() time.Duration {
	duration, err := str2duration.ParseDuration(ac.TokenAccessValidity)
	if err != nil {
		log.Logf(log.LevelCritical, "unable to parse TokenAccessValidity duration: %s, assuming default %s", ac.TokenAccessValidity, TOKEN_ACCESS_VALIDITY.String())
		return TOKEN_ACCESS_VALIDITY
	}
	return duration
}

func (ac ApiConfig) TokenRefreshValidityDuration() time.Duration {
	duration, err := str2duration.ParseDuration(ac.TokenAccessValidity)
	if err != nil {
		log.Logf(log.LevelCritical, "unable to parse TokenRefreshValidity duration: %s, assuming default %s", ac.TokenRefreshValidity, TOKEN_REFRESH_VALIDITY.String())
		return TOKEN_REFRESH_VALIDITY
	}
	return duration
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

// intentionally obfuscated json keys for security and bandwidth savings
type JwtRole struct {
	OrgView  bool `json:"a"`
	OrgEdit  bool `json:"b"`
	OrgAdmin bool `json:"c"`
}

type JwtRoles map[int]JwtRole

func JwtRolesFromTable(roles []tables.UserRoles) JwtRoles {
	jwtRoles := JwtRoles{}
	for _, r := range roles {
		jwtRoles[r.OrgID] = JwtRole{
			OrgView:  r.OrgView,
			OrgEdit:  r.OrgEdit,
			OrgAdmin: r.OrgAdmin,
		}
	}
	return jwtRoles
}

// func (roles JwtRoles) GetTableUserRoles(userId int) []tables.UserRoles {
// 	var userRoles []tables.UserRoles
// 	for orgID, r := range roles {
// 		userRoles = append(userRoles, tables.UserRoles{
// 			UserID:   userId,
// 			OrgID:    orgID,
// 			OrgView:  r.OrgView,
// 			OrgEdit:  r.OrgEdit,
// 			OrgAdmin: r.OrgAdmin,
// 		})
// 	}
// 	return userRoles
// }

// If changes are made to JwtAccessClaims, this revision uint must be incremented
const LatestAccessTokenRevision uint = 1

// intentionally obfuscated json keys for security and bandwidth savings
type JwtAccessClaims struct {
	UserID     int      `json:"a"`
	Roles      JwtRoles `json:"b"`
	SuperAdmin bool     `json:"c"`
	CSRFHeader string   `json:"d"`
	Revision   uint     `json:"e"`
	jwt.RegisteredClaims
}

func CreateJwtAccessClaims(userID int, roles JwtRoles, superAdmin bool, csrfHeader string) *JwtAccessClaims {
	return &JwtAccessClaims{
		UserID:     userID,
		Roles:      roles,
		SuperAdmin: superAdmin,
		CSRFHeader: csrfHeader,
		Revision:   LatestAccessTokenRevision,
	}
}

// If changes are made to JwtAccessClaims, this revision uint must be incremented
const LatestRefreshTokenRevision uint = 1

// intentionally obfuscated json keys for security and bandwidth savings
type JwtRefreshClaims struct {
	UserID     int    `json:"a"`
	Nonce      string `json:"b"`
	CSRFHeader string `json:"c"` // TODO: maybe remove CSRF Token from access or refresh claim to reduce bandwidth usage
	Revision   uint   `json:"d"`
	jwt.RegisteredClaims
}

func CreateJwtRefreshClaims(userID int, nonce string, csrfHeader string) *JwtRefreshClaims {
	return &JwtRefreshClaims{
		UserID:     userID,
		Nonce:      nonce,
		CSRFHeader: csrfHeader,
		Revision:   LatestRefreshTokenRevision,
	}
}

// type HandleFunc func(c echo.Context) error
// type HandleFunc echo.HandlerFunc

type StateReferrer struct {
	Nonce string `json:"nonce"`
	URI   string `json:"uri"`
}
