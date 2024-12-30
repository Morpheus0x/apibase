package db

import "gopkg.cc/apibase/log"

var (
	ErrDatabaseConfig    = log.RegisterErrType("database config invalid")
	ErrDatabaseMigration = log.RegisterErrType("database migration failed")
	ErrDatabaseConn      = log.RegisterErrType("database connect failed")
	ErrDatabaseQuery     = log.RegisterErrType("database query error")
	ErrDatabaseNotFound  = log.RegisterErrType("database entry not found")
	ErrDatabaseCommit    = log.RegisterErrType("database tx commit failed")
	ErrDatabaseScan      = log.RegisterErrType("database scan to struct failed")
)
