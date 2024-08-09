package db

import (
	"fmt"
	"reflect"

	"gopkg.cc/apibase/sqlite"
	"gorm.io/gorm"
)

// Generic db driver for package web

type DBKind uint

const (
	SQLite DBKind = iota
	PostgreSQL
)

type DBAdapter uint

const (
	Apibase DBAdapter = iota
	Gorm
	Pgx
)

type DB struct {
	Kind     DBKind
	Adapter  DBAdapter
	SQLite   *sqlite.SQLite
	Postgres any
}

func RawQuery[T any](db *DB, query string) ([]*T, error) {
	switch db.Kind {
	case SQLite:
		if db.Adapter != Apibase {
			return []*T{}, fmt.Errorf("invalid sqlite adapter")
		}
		return sqlite.RawQuery[T](db.SQLite, query)
	case PostgreSQL:
		if db.Adapter != Gorm {
			return []*T{}, fmt.Errorf("only gorm is implemented for postgres")
		}
		if reflect.TypeOf(db.Postgres) != reflect.TypeOf(&gorm.DB{}) {
			return []*T{}, fmt.Errorf("db.Postgres must be valid gorm adapter")
		}
		model := []T{}
		tx := db.Postgres.(*gorm.DB).Model(&model).Raw(query)
		if tx.Error != nil {
			return []*T{}, fmt.Errorf("error running raw postgres query: %v", tx.Error)
		}
		out := []*T{}
		for _, t := range model {
			out = append(out, &t)
		}
		return out, nil
	default:
		return []*T{}, fmt.Errorf("no valid DB specified")
	}
}
