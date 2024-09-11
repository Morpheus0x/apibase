package db

import (
	"fmt"

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

func ValidateDB(database DB) error {
	switch database.Kind {
	case SQLite:
		if database.SQLite == nil {
			return fmt.Errorf("no valid SQLite database adapter")
		}
	case PostgreSQL:
		if database.Gorm == nil {
			return fmt.Errorf("no valid PostgreSQL database adapter")
		}
	default:
		return fmt.Errorf("no valid DB specified")
	}
	return nil
}

func MigrateDefaultTables(database DB) error {
	switch database.Kind {
	case SQLite:
		// TODO: do this
		return fmt.Errorf("WIP, not implemented")
	case PostgreSQL:
		err := database.Gorm.AutoMigrate(&Users{})
		if err != nil {
			return err
		}
		err = database.Gorm.AutoMigrate(&RefreshTokens{})
		if err != nil {
			return err
		}
		// TODO: use own logger
		fmt.Printf("Successfully migrated PostgreSQL Tables.")
	default:
		return fmt.Errorf("no valid DB specified")
	}
	return nil
}
