package db

import (
	"fmt"

	"gopkg.cc/apibase/sqlite"
)

func RawQuery[T any](db *DB, query string) ([]*T, error) {
	switch db.Kind {
	case SQLite:
		return sqlite.RawQuery[T](db.SQLite, query)
	case PostgreSQL:
		model := []T{}
		tx := db.Gorm.Model(&model).Raw(query)
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
