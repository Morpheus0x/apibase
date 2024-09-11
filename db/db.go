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
