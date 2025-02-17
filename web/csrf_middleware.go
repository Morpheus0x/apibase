package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	h "gopkg.cc/apibase/helper"
	"gopkg.cc/apibase/log"
	wr "gopkg.cc/apibase/web_response"
)

// echo middleware to verify valid CSRF header and cookie
func CheckCSRF(api *ApiServer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Log(log.LevelDevel, "Middlware CheckCSRF executed")
			csrfHeader := c.Request().Header.Get("X-XSRF-TOKEN")
			csrfCookie, err := c.Request().Cookie("csrf_token")
			if err != nil || csrfHeader != csrfCookie.Value {
				return wr.SendJsonErrorResponse(c, http.StatusForbidden, wr.RespErrCsrfInvalid)
			}
			return next(c)
		}
	}
}

// echo get endpoint to ask for new CSRF
func GetCSRF(api *ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := c.Request().Cookie("csrf_token")
		if err != nil {
			UpdateCSRF(c, api, h.CreateSecretString(""))
		}
		return nil
	}
}

// update
func UpdateCSRF(c echo.Context, api *ApiServer, sessionId h.SecretString) {
	var csrfToken h.SecretString
	if sessionId.GetSecret() == "" {
		csrfToken = createCSRF(api, getSessionId(c, api.Config.TokenSecretBytes()))
	} else {
		csrfToken = createCSRF(api, h.CreateSecretString(""))
	}

	expiresIn := api.Config.AddCookieExpiryMargin(api.Config.Settings.TokenRefreshValidity)
	csrfCookie := &http.Cookie{Name: "csrf_token", Value: csrfToken.GetSecret(), Path: "/", Expires: time.Now().Add(expiresIn)}
	h.OverwriteRequestCookie(c.Request(), csrfCookie) // set cookie for current request
	c.SetCookie(csrfCookie)                           // set cookie for response
}

func createCSRF(api *ApiServer, sessionId h.SecretString) h.SecretString {
	hash := hmac.New(sha256.New, api.Config.TokenSecretBytes())
	hash.Write([]byte(h.RandomString(16) + sessionId.GetSecret()))
	return h.CreateSecretString(base64.RawURLEncoding.EncodeToString(hash.Sum(nil)))
}

func getSessionId(c echo.Context, tokenSecret []byte) h.SecretString {
	refreshToken, err := parseRefreshTokenCookie(c, tokenSecret)
	if err != nil {
		return h.CreateSecretString("")
	}
	refreshClaims, ok := refreshToken.Claims.(*jwtRefreshClaims)
	if !ok {
		return h.CreateSecretString("")
	}
	return refreshClaims.SessionID
}

// func RemoveCSRF(c echo.Context) {
// 	c.SetCookie(&http.Cookie{Name: "csrf_token", Value: "", Path: "/", Expires: time.Unix(0, 0)})
// }
