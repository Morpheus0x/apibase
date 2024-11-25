package web

import "github.com/labstack/echo/v4"

func oauthLogin(api *ApiServer) echo.HandlerFunc {
	return func(c echo.Context) error {
		request := c.Request()
		queryURL := request.URL.Query()
		queryURL.Set("provider", c.Param("provider"))
		queryURL.Set("state", "random_var_for_csrf_and_return_uri")
		request.URL.RawQuery = queryURL.Encode()

		// gothic.BeginAuthHandler
		return nil
	}
}
