package web_setup

import "gopkg.cc/apibase/errx"

var (
	ErrWebBind          = errx.NewType("invalid bind")
	ErrWebApiNotInit    = errx.NewType("ApiServer not initialized")
	ErrWebUnknownMethod = errx.NewType("Unknown HTTP Method")
	ErrWebShutdown      = errx.NewType("unable to gracefully shutdown web server")
)
