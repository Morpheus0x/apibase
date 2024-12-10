package db

import "gopkg.cc/apibase/log"

var (
	ErrDatabaseConfig    = log.RegisterErrType("database config invalid")
	ErrDatabaseMigration = log.RegisterErrType("database migration failed")
	ErrDatabaseConn      = log.RegisterErrType("database connect failed")
)
