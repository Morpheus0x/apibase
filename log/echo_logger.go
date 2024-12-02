package log

import (
	"fmt"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/color"
)

func EchoLoggerMiddleware(level Level) echo.MiddlewareFunc {
	loggerMutex.RLock()
	loggerLocal := logger
	loggerMutex.RUnlock()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			col := color.New()

			statusRaw := res.Status
			status := col.Green(statusRaw)
			switch {
			case statusRaw >= 500:
				status = col.Red(statusRaw)
			case statusRaw >= 400:
				status = col.Yellow(statusRaw)
			case statusRaw >= 300:
				status = col.Cyan(statusRaw)
			}
			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			latency := stop.Sub(start)

			// echo default logger format
			// `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			// 	`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			// 	`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			// 	`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
			out := fmt.Sprintf("HTTP %s %s, ID: %s, IP: %s, URI: %s, Latency: %s (%s)",
				req.Method,
				status,
				id,
				c.RealIP(),
				req.RequestURI,
				strconv.FormatInt(int64(latency), 10),
				latency.String(),
			)

			for _, w := range loggerLocal.Writers {
				var logOutput []byte
				if w.WithTime {
					logOutput = []byte(fmt.Sprintf("%s %s %s\n", time.Now().Format(loggerLocal.TimeFmt), level.String(), out))
				} else {
					logOutput = []byte(fmt.Sprintf("%s %s\n", level.String(), out))
				}
				w.Lock()
				w.Writer.Write(logOutput)
				w.Unlock()
			}
			return nil
		}
	}
}
