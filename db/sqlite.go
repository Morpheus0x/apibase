package db

import (
	"gopkg.cc/apibase/baseconfig"
	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/sqlite"
)

func InitSQLite(config SQLiteConfig, bc baseconfig.BaseConfig) (*sqlite.SQLite, error) {
	// TODO: validate path and use LockFile
	sqlite, err := sqlite.OpenWithConfig(config.FilePath, sqlite.SQLiteConfig{SQLITE_DATETIME_FORMAT: bc.SQLiteDatetimeFormat})
	if err != nil {
		return sqlite, errx.WrapWithType(ErrDatabaseConn, err, "unable to open sqlite database")
	}
	return sqlite, nil
}
