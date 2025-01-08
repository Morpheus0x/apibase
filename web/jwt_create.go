package web

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func createJwtAccessClaims(userID int, roles JwtRoles, superAdmin bool, csrfHeader string) *JwtAccessClaims {
	return &JwtAccessClaims{
		UserID:     userID,
		Roles:      roles,
		SuperAdmin: superAdmin,
		CSRFHeader: csrfHeader,
		Revision:   LatestAccessTokenRevision,
	}
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

func createJwtRefreshClaims(userID int, nonce string, csrfHeader string) *JwtRefreshClaims {
	return &JwtRefreshClaims{
		UserID:     userID,
		Nonce:      nonce,
		CSRFHeader: csrfHeader,
		Revision:   LatestRefreshTokenRevision,
	}
}
