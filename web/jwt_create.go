package web

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	h "gopkg.cc/apibase/helper"
)

// Access Token

// If changes are made to JwtAccessClaims, this revision uint must be incremented
const LatestAccessTokenRevision uint = 1

// intentionally obfuscated json keys for security and bandwidth savings
type jwtAccessClaims struct {
	UserID     int      `json:"a"`
	Roles      JwtRoles `json:"b"`
	SuperAdmin bool     `json:"c"`
	Revision   uint     `json:"d"`
	jwt.RegisteredClaims
}

func (claims *jwtAccessClaims) signToken(api *ApiServer) (string, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(api.Config.Settings.TokenAccessValidity))
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return rawToken.SignedString(api.Config.TokenSecretBytes())
}

func createJwtAccessClaims(userID int, roles JwtRoles, superAdmin bool) *jwtAccessClaims {
	return &jwtAccessClaims{
		UserID:     userID,
		Roles:      roles,
		SuperAdmin: superAdmin,
		Revision:   LatestAccessTokenRevision,
	}
}

// Refresh Token

// If changes are made to JwtAccessClaims, this revision uint must be incremented
const LatestRefreshTokenRevision uint = 1

// intentionally obfuscated json keys for security and bandwidth savings
type jwtRefreshClaims struct {
	UserID    int            `json:"a"`
	SessionID h.SecretString `json:"b"`
	Revision  uint           `json:"c"`
	jwt.RegisteredClaims
}

func (claims *jwtRefreshClaims) signToken(api *ApiServer) (string, time.Time, error) {
	now := time.Now()
	claims.IssuedAt = jwt.NewNumericDate(now)
	expiresAt := now.Add(api.Config.Settings.TokenRefreshValidity)
	claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	token, err := rawToken.SignedString(api.Config.TokenSecretBytes())
	return token, expiresAt, err

}

func createJwtRefreshClaims(userID int, sessionId h.SecretString) *jwtRefreshClaims {
	return &jwtRefreshClaims{
		UserID:    userID,
		SessionID: sessionId,
		Revision:  LatestRefreshTokenRevision,
	}
}
