package webtype

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/log"
)

// Access Token

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

func (claims *JwtAccessClaims) SignToken(api *ApiServer) (string, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(api.Config.TokenAccessValidityDuration()))
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return rawToken.SignedString(api.Config.TokenSecretBytes())
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

func ParseAccessTokenCookie(c echo.Context, secret []byte) (*jwt.Token, *log.Error) {
	tokenRaw, err := c.Cookie("access_token")
	if err != nil {
		// c.Logger().Debugf("no cookie 'access_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "no cookie 'access_token' present in request")
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(JwtAccessClaims), func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		// c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "error parsing token 'access_token'")
	}
	return token, log.ErrorNil()
}

// Refresh Token

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

func (claims *JwtRefreshClaims) SignToken(api *ApiServer) (string, time.Time, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	expiresAt := now.Add(api.Config.TokenRefreshValidityDuration())
	claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	token, err := rawToken.SignedString(api.Config.TokenSecretBytes())
	return token, expiresAt, err

}

func CreateJwtRefreshClaims(userID int, nonce string, csrfHeader string) *JwtRefreshClaims {
	return &JwtRefreshClaims{
		UserID:     userID,
		Nonce:      nonce,
		CSRFHeader: csrfHeader,
		Revision:   LatestRefreshTokenRevision,
	}
}

func ParseRefreshTokenCookie(c echo.Context, secret []byte) (*jwt.Token, *log.Error) {
	tokenRaw, err := c.Cookie("refresh_token")
	if err != nil {
		// c.Logger().Debugf("no cookie 'refresh_token' in request (%s): %v", c.Request().RequestURI, err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "no cookie 'refresh_token' present in request")
	}
	token, err := jwt.ParseWithClaims(tokenRaw.Value, new(JwtRefreshClaims), func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		// c.Logger().Debugf("error parsing token from cookie: %v", err)
		return &jwt.Token{}, log.NewErrorWithType(ErrTokenValidate, "error parsing token 'refresh_token'")
	}
	return token, log.ErrorNil()
}
