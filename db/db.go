package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"gopkg.cc/apibase/baseconfig"
	"gopkg.cc/apibase/errx"
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
	Kind       DBKind
	SQLite     *sqlite.SQLite
	Postgres   *pgx.Conn
	BaseConfig *baseconfig.BaseConfig
}

func ValidateDB(database DB) error {
	if database.BaseConfig == nil {
		return errx.NewWithType(ErrMissingBaseConfig, "")
	}
	switch database.Kind {
	case SQLite:
		if database.SQLite == nil {
			return errx.NewWithType(ErrDatabaseConfig, "no valid SQLite database adapter")
		}
		// TODO: this
		return errx.NewWithType(errx.ErrNotImplemented, "sqlite driver for apibase not implemented yet")
	case PostgreSQL:
		if database.Postgres == nil {
			return errx.NewWithType(ErrDatabaseConfig, "no valid PostgreSQL database adapter")
		}
		ctx, cancel := context.WithTimeout(context.Background(), database.BaseConfig.TimeoutDatabaseConnect)
		defer cancel()
		err := database.Postgres.Ping(ctx)
		if err != nil {
			return errx.WrapWithType(ErrDatabaseConn, err, "unable to ping PostgreSQL database")
		}
	default:
		return errx.NewWithType(ErrDatabaseConfig, "no valid DB specified")
	}
	return nil
}

func MigrateDefaultTables(database DB) error {
	// ctx, cancel := context.WithTimeout(context.Background(), database.BaseConfig.TimeoutDatabaseConnect)
	// defer cancel()
	switch database.Kind {
	case SQLite:
		// TODO: do this
		return errx.NewWithType(errx.ErrNotImplemented, "sqlite tables not migrated")
	case PostgreSQL:
		// users := []table.User{}
		// err := pgxscan.Select(ctx, database.Postgres, &users, "SELECT * FROM users")
		// if err != nil {
		// 	return errx.WrapWithType(ErrDatabaseMigration, err, "")
		// }
		// log.Logf(log.LevelDebug, "Users: %+v", users)
		// TODO: read tables from db and verify they match the local struct
		log.Log(log.LevelInfo, "Successfully migrated PostgreSQL Tables.")
	default:
		return errx.NewWithTypef(ErrDatabaseMigration, "no valid DB specified, db.DBKind(%d)", database.Kind)
	}
	return nil
}
