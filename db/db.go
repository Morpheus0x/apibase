package db

import (
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/sqlite"
	"gorm.io/gorm"
)

// Generic db driver for package web

type DBKind uint

const (
	SQLite DBKind = iota
	PostgreSQL
)

type DB struct {
	Kind   DBKind
	SQLite *sqlite.SQLite
	Gorm   *gorm.DB
}

func ValidateDB(database DB) *log.Error {
	switch database.Kind {
	case SQLite:
		if database.SQLite == nil {
			return log.NewErrorWithType(ErrDatabaseConfig, "no valid SQLite database adapter")
		}
	case PostgreSQL:
		if database.Gorm == nil {
			return log.NewErrorWithType(ErrDatabaseConfig, "no valid PostgreSQL database adapter")
		}
	default:
		return log.NewErrorWithType(ErrDatabaseConfig, "no valid DB specified")
	}
	return nil
}

func MigrateDefaultTables(database DB) *log.Error {
	switch database.Kind {
	case SQLite:
		// TODO: do this
		return log.NewErrorWithType(ErrDatabaseMigration, "WIP, not implemented")
	case PostgreSQL:
		err := database.Gorm.AutoMigrate(&Users{})
		if err != nil {
			return log.NewErrorWithType(ErrDatabaseMigration, err.Error())
		}
		err = database.Gorm.AutoMigrate(&RefreshTokens{})
		if err != nil {
			return log.NewErrorWithType(ErrDatabaseMigration, err.Error())
		}
		log.Log(log.LevelInfo, "Successfully migrated PostgreSQL Tables.")
	default:
		return log.NewErrorWithTypef(ErrDatabaseMigration, "no valid DB specified, db.DBKind(%d)", database.Kind)
	}
	return nil
}
