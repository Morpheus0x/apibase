package db

import "gopkg.cc/apibase/errx"

var (
	ErrDatabaseConfig    = errx.NewType("database config invalid")
	ErrDatabaseMigration = errx.NewType("database migration failed")
	ErrDatabaseConn      = errx.NewType("database connect failed")
	ErrDatabaseQuery     = errx.NewType("database query error")
	ErrDatabaseNotFound  = errx.NewType("database entry not found")
	ErrDatabaseCommit    = errx.NewType("database tx commit failed")
	ErrDatabaseScan      = errx.NewType("database scan to struct failed")
	ErrDatabaseInsert    = errx.NewType("database insert into failed")
	ErrDatabaseUpdate    = errx.NewType("database update failed")
	ErrDatabaseDelete    = errx.NewType("database delete failed")
	ErrUserAlreadyExists = errx.NewType("user already exists")
	ErrNoRoles           = errx.NewType("missing required roles")
	ErrOrgCreate         = errx.NewType("organization couldn't be created")
	ErrMissingBaseConfig = errx.NewType("baseconfig.BaseConfig is not defined for db.DB")
)
