package web

import "gopkg.cc/apibase/log"

var (
	ErrWebBind          = log.RegisterErrType("invalid bind")
	ErrWebApiNotInit    = log.RegisterErrType("ApiServer not initialized")
	ErrWebUnknownMethod = log.RegisterErrType("Unknown HTTP Method")
	ErrWebShutdown      = log.RegisterErrType("unable to gracefully shutdown web server")
)
