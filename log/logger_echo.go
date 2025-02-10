package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/color"
)

func EchoLoggerMiddleware(defaultLevel Level, apiRootLevel Level) echo.MiddlewareFunc {
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

			level := defaultLevel
			path := c.Path()
			if !strings.HasPrefix(path, "/api") && !strings.HasPrefix(path, "/auth") {
				level = apiRootLevel
			}
			if level < loggerLocal.Level {
				// log message is too verbose, level is higher than set in logger, ignoring
				return
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
			// id := req.Header.Get(echo.HeaderXRequestID)
			// if id == "" {
			// 	id = res.Header().Get(echo.HeaderXRequestID)
			// }

			// echo default logger format
			// `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			// 	`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			// 	`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
			// 	`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
			out := fmt.Sprintf("HTTP %s %s, IP: %s, URI: %s, Latency: %s", // , ID: %s
				req.Method,
				status,
				// id,
				c.RealIP(),
				req.RequestURI,
				stop.Sub(start).String(),
			)

			for _, w := range loggerLocal.Writers {
				logOutput := loggerLocal.formatLogOutput(level, out, loggerLocal.TimeFmt, w.WithTime)
				w.Lock()
				w.Writer.Write(logOutput)
				w.Unlock()
			}
			return nil
		}
	}
}

func EchoLogPanicStacktrace(c echo.Context, err error, stack []byte) error {
	Logf(LevelDebug, "[PANIC RECOVER] Error: %v, Stacktrace: %s", err, stack)
	return nil
}
