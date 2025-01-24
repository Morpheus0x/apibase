package cron

import "gopkg.cc/apibase/errx"

var (
	ErrTaskExists       = errx.NewType("task already exists")
	ErrTaskStartup      = errx.NewType("unable to start task")
	ErrTaskStartTimeout = errx.NewType("task took too long to startup")
)
