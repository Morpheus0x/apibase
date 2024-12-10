package db

import (
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/sqlite"
)

func InitSQLite(config SQLiteConfig, bc sqlite.SQLiteConfig) (*sqlite.SQLite, *log.Error) {
	// TODO: validate path and use LockFile
	sqlite, err := sqlite.OpenWithConfig(config.FilePath, bc)
	if err != nil {
		return sqlite, log.NewErrorWithTypef(ErrDatabaseConn, "unable to open sqlite database: %s", err.Error())
	}
	return sqlite, log.ErrorNil()
}
