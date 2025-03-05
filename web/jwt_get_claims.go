package web

import (
	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/errx"
)

// Get access claims with optional custom data.
// To correctly parse access claim data, initialize empty struct of correct type using: api.GetAccessClaimDataType() or new(<your_custom_struct_type>)
func GetAccessClaims[T any](c echo.Context, api *ApiServer, data *T) (*jwtAccessClaims[T], error) {
	accessToken, err := parseAccessTokenCookie(c, api.Config.TokenSecretBytes(), data)
	if err != nil {
		return &jwtAccessClaims[T]{}, err
	}
	accessClaims, ok := accessToken.Claims.(*jwtAccessClaims[T])
	if !ok {
		return accessClaims, errx.NewWithType(ErrAccessClaimsParsing, "")
	}
	if accessClaims.Data == nil {
		return accessClaims, errx.NewWithType(ErrAccessClaimDataNil, "")
	}
	return accessClaims, nil
}
