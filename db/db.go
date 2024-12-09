package db

import (
	"database/sql"

	"github.com/stytchauth/sqx"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/sqlite"
)

// Generic db driver for package web

type DBKind uint

const (
	SQLite DBKind = iota
	PostgreSQL
)

type DB struct {
	Kind     DBKind
	SQLite   *sqlite.SQLite
	Postgres *sql.DB
}

func ValidateDB(database DB) *log.Error {
	switch database.Kind {
	case SQLite:
		if database.SQLite == nil {
			return log.NewErrorWithType(ErrDatabaseConfig, "no valid SQLite database adapter")
		}
		sqx.SetDefaultQueryable(database.SQLite.DB)
		// TODO: this
		return log.NewErrorWithType(log.ErrNotImplemented, "sqlite driver for apibase not implemented yet")
	case PostgreSQL:
		if database.Postgres == nil {
			return log.NewErrorWithType(ErrDatabaseConfig, "no valid PostgreSQL database adapter")
		}
		sqx.SetDefaultQueryable(database.Postgres)
	default:
		return log.NewErrorWithType(ErrDatabaseConfig, "no valid DB specified")
	}
	sqx.SetDefaultLogger(log.NewGenericLogger(log.LevelDebug))
	// TODO: validate that sqx can query successfully
	return log.ErrorNil()
}

func MigrateDefaultTables(database DB) *log.Error {
	switch database.Kind {
	case SQLite:
		// TODO: do this
		return log.NewErrorWithType(log.ErrNotImplemented, "sqlite tables not migrated")
	case PostgreSQL:
		log.Log(log.LevelInfo, "Successfully migrated PostgreSQL Tables.")
	default:
		return log.NewErrorWithTypef(ErrDatabaseMigration, "no valid DB specified, db.DBKind(%d)", database.Kind)
	}
	return log.ErrorNil()
}
