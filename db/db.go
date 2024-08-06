package db

import "gopkg.cc/apibase/sqlite"

// Generic db driver for package web

type DBKind uint

const (
	SQLite DBKind = iota
	PostgreSQL
)

type DB struct {
	kind     DBKind
	sqlite   *sqlite.SQLite
	postgres string // TODO: new
}
