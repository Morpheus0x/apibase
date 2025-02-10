package web

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"gopkg.cc/apibase/log"
	wr "gopkg.cc/apibase/web_response"
)

func EchoErrorHandler(err error, c echo.Context) {
	if err == nil || c.Response().Committed {
		return
	}
	status := http.StatusInternalServerError
	body := wr.JsonResponse[struct{}]{}

	// Parse error
	if he, ok := err.(*echo.HTTPError); ok {
		if he.Internal != nil {
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				he = herr
			}
		}
		status = he.Code
		switch m := he.Message.(type) {
		case string:
			log.Logf(log.LevelDebug, "HTTP Msg: %s, Error: %s", m, err.Error())
			body.Message = m
		case json.Marshaler:
			// do nothing - this type knows how to format itself to JSON
		case error:
			body.Message = m.Error()
		}
	} else if re, ok := err.(*wr.ResponseError); ok {
		body, status = re.GetResponse()
	} else {
		log.Logf(log.LevelWarning, "Unknown error for request '%s %s' (%s): %s", c.Request().Method, c.Request().URL.Path, wr.RespErrUnknownInternal.String(), err.Error())
		body = wr.JsonResponse[struct{}]{
			ErrorID: wr.RespErrUnknownInternal,
		}
	}

	// Send response
	if c.Request().Method == http.MethodHead {
		err = c.NoContent(status)
	} else {
		err = c.JSON(status, body)
	}
	if err != nil {
		log.Logf(log.LevelError, "Unable to send response in echo error handler: %s", err.Error())
	}
}
