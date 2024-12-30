package db

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/sqlite"
	"gopkg.cc/apibase/tables"
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
	Postgres *pgx.Conn
}

func ValidateDB(database DB) *log.Error {
	switch database.Kind {
	case SQLite:
		if database.SQLite == nil {
			return log.NewErrorWithType(ErrDatabaseConfig, "no valid SQLite database adapter")
		}
		// TODO: this
		return log.NewErrorWithType(log.ErrNotImplemented, "sqlite driver for apibase not implemented yet")
	case PostgreSQL:
		if database.Postgres == nil {
			return log.NewErrorWithType(ErrDatabaseConfig, "no valid PostgreSQL database adapter")
		}
	default:
		return log.NewErrorWithType(ErrDatabaseConfig, "no valid DB specified")
	}
	return log.ErrorNil()
}

func MigrateDefaultTables(database DB) *log.Error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // TODO: remove hardcoded timeout
	defer cancel()
	switch database.Kind {
	case SQLite:
		// TODO: do this
		return log.NewErrorWithType(log.ErrNotImplemented, "sqlite tables not migrated")
	case PostgreSQL:
		users := []tables.Users{}
		err := pgxscan.Select(ctx, database.Postgres, &users, "SELECT * FROM users")
		if err != nil {
			return log.NewErrorWithType(ErrDatabaseMigration, err.Error())
		}
		log.Logf(log.LevelDebug, "Users: %+v", users)
		// TODO: read tables from db and verify they match the local struct
		log.Log(log.LevelInfo, "Successfully migrated PostgreSQL Tables.")
	default:
		return log.NewErrorWithTypef(ErrDatabaseMigration, "no valid DB specified, db.DBKind(%d)", database.Kind)
	}
	return log.ErrorNil()
}
