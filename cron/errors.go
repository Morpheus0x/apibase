package cron

import "gopkg.cc/apibase/errx"

var (
	ErrTaskExists         = errx.NewType("task already exists")
	ErrTaskStartup        = errx.NewType("unable to start task")
	ErrTaskStartTimeout   = errx.NewType("task took too long to startup")
	ErrTaskRemove         = errx.NewType("unable to remove task")
	ErrTaskNoRunFunc      = errx.NewType("task to schedule doesn't have a function to run")
	ErrTaskDatabaseSave   = errx.NewType("unable to save task to database")
	ErrTaskDatabaseDelete = errx.NewType("unable to delete task from database")
)
