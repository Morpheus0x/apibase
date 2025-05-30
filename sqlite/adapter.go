package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// type Query struct {
// 	Table      string
// 	Column     string
// 	Value      interface{}
// 	Conditions interface{}
// }

// type rawQuery struct {
// 	Table Table
// 	Query string
// }

// TODO: make all errors global vars

type SQLite struct {
	path   string
	tables []Table // TODO: register table and match on type of out in Query func
	// tableMap map[string]Table
	DB *sql.DB

	startTime int64
	config    SQLiteConfig
}

type SQLiteConfig struct {
	SQLITE_DATETIME_FORMAT string
}

func Open(path string) (*SQLite, error) {
	var err error
	// TODO: maybe validate path?
	sqlite := &SQLite{
		path:      path,
		startTime: time.Now().Unix(),
		config: SQLiteConfig{
			SQLITE_DATETIME_FORMAT: "2006-01-02 15:04:05",
		},
	}
	sqlite.DB, err = sql.Open("sqlite3", sqlite.path)
	return sqlite, err
}

func OpenWithConfig(path string, conf SQLiteConfig) (*SQLite, error) {
	var err error
	// TODO: maybe validate path?
	sqlite := &SQLite{
		path:      path,
		startTime: time.Now().Unix(),
		config:    conf,
	}
	sqlite.DB, err = sql.Open("sqlite3", sqlite.path)
	return sqlite, err
}

func (s *SQLite) Close() error {
	return s.DB.Close()
}

func (s *SQLite) tableExists(name string) bool {
	for _, table := range s.tables {
		if table.Name == name {
			return true
		}
	}
	return false
}

func (s *SQLite) TableMust(name string, data any) {
	err := s.Table(name, data)
	if err != nil {
		panic(err)
	}
}

// Create table or migrates existing one
// Requirements:
// Adding columns is possible, but removing or changing type or attributes is not
// new unique constrains must be added below the existing ones
// IMPORTANT: The primary key must always be named id, e.g. `db:"id"...`
func (s *SQLite) Table(name string, data any) error {
	// TODO: add old_name for table rename migrations ?
	if s.tableExists(name) {
		return fmt.Errorf("table with same name already exists")
	}

	table := Table{Name: name, Data: data}
	// fmt.Printf("table: %v\ndata: %v\n", table, data)
	// Query existing tables
	var existingSchema string
	err := s.DB.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name=?;", name).Scan(&existingSchema)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("unabel to query for existing table: %v", err)
	}
	if err == sql.ErrNoRows {
		// Create Table if it doesn't exist yet
		schema, err := table.Schema()
		if err != nil {
			return fmt.Errorf("unable to get schema of table '%s': %v", name, err)
		}
		_, err = s.DB.Exec(schema)
		if err != nil {
			return fmt.Errorf("unable to create table '%s': %v, schema: %s", name, err, schema)
		}
		fmt.Printf("### New table created\n%s:\n", schema)
		s.tables = append(s.tables, table)
		// fmt.Printf("return here why? schema: %s\n res: %v", schema, res)
		return nil
	}
	// fmt.Printf("existingSchema: %s", existingSchema)
	err = s.deployUpdatedTable(table, existingSchema)
	if err != nil {
		return fmt.Errorf("unable to migrate existing table: %v", err)
	}

	s.tables = append(s.tables, table)
	// s.tableMap[name] = table

	return nil
}

func (s *SQLite) deployUpdatedTable(t Table, existingTableSchema string) error {
	newSchema, err := t.Schema()
	if err != nil {
		return err
	}
	if newSchema == existingTableSchema {
		return nil
	}
	fmt.Printf("### Existing table doesn't match created schema from struct:\n")
	fmt.Printf("### New Table:\n%v\n", newSchema)
	fmt.Printf("### Existing Table:\n%v\n", existingTableSchema)

	backupTableName := fmt.Sprintf("%s_bak%d", t.Name, s.startTime)

	// Rename existing DB which old schema
	res, err := s.DB.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", t.Name, backupTableName))
	if err != nil {
		return fmt.Errorf("unable to create table backup: %v, res: %v", err, res)
	}
	// Create new target DB
	_, err = s.DB.Exec(newSchema)
	if err != nil {
		return fmt.Errorf("unable to create updated table: %v", err)
	}
	// Copy data to new table
	cols, err := Table{Name: t.Name, raw_schema: existingTableSchema}.ColumnQueryString()
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s;", t.Name, cols, cols, backupTableName))
	if err != nil {
		migrateError := fmt.Errorf("unable to copy data from backup to updated table: %v", err)
		s.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS  %s;", t.Name))
		_, revertError := s.DB.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", backupTableName, t.Name))
		if revertError != nil {
			return fmt.Errorf("unable to restore table after failed migration: %v, original error: %v", revertError, migrateError)
		}
		return migrateError
	}
	// Delete backup table with old schema
	// db.Exec("DROP TABLE ?", backupTableName);
	return nil
}

func (t Table) ColumnQueryString() (string, error) {
	cols, err := t.Columns(true)
	if err != nil {
		return "", err
	}
	return strings.Join(cols, ", "), nil // maybe use cols[:] instead to convert array to slice?? Source: https://stackoverflow.com/a/28799151
}

func (t Table) Columns(nameOnly bool) ([]string, error) {
	out := []string{}
	schema, err := t.Schema()
	if err != nil {
		return out, err
	}
	lines := strings.Split(schema, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "CREATE TABLE") || strings.HasPrefix(line, ")") {
			continue
		}
		if nameOnly {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "UNIQUE") {
				continue
			}
			out = append(out, strings.Split(trimmed, " ")[0])
		} else {
			out = append(out, line)
		}
	}
	return out, nil
}

// func (s *SQLite) parseTable(data any) error {

// }

// func (s *SQLite) addOrMigrateTable(table Table)
