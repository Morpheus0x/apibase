package sqlite

import (
	"fmt"
	"reflect"
	"strings"
)

// return string: where clause, MUST start with a space
type Builder func(data any) (string, error)

// User can pass a function to make custom queryies (where clause)
// builder: user defined query function
// data: custom data used to build where clause
// out: this determines what table to query and will be used as data return
func (s *SQLite) QueryWithBuilder(builder Builder, data any, out *[]any) error { // Where
	where, err := builder(data)
	if err != nil {
		return err
	}
	// TODO: check if first char of where is a space, if not, return error
	table, err := s.getTable(out)
	if err != nil {
		return err
	}
	columns, err := table.ColumnQueryString()
	if err != nil {
		return err
	}
	query := fmt.Sprintf("SELECT %s FROM %s%s;", columns, table.Name, where)
	outData, err := s.sqliteQuery(table, query)
	out = &outData
	return err
}

// full SQL query without FROM ...
func (s *SQLite) Query(query string, out *[]any) error { // Raw
	table, err := s.getTable(out)
	if err != nil {
		return err
	}
	nextIsTableName := false
	for _, s := range strings.Split(query, " ") {
		if nextIsTableName {
			if s != table.Name {
				return fmt.Errorf("table name in query doesn't match table from out datatype")
			}
			break
		}
		if s == "FROM" {
			nextIsTableName = true
		}
	}

	outData, err := s.sqliteQuery(table, query)
	out = &outData
	return err
}

func (s *SQLite) sqliteQuery(table Table, query string) ([]interface{}, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return []interface{}{nil}, err
	}
	out, err := s.scanRows(rows, reflect.TypeOf(table.Data))
	return out, err
}

// TODO: shoud this return Table or string with table name?
func (s *SQLite) getTable(out *[]any) (Table, error) {
	for _, t := range s.tables {
		if reflect.TypeOf(t.Data) == reflect.TypeOf(*out).Elem() {
			return t, nil
		}
	}
	return Table{}, fmt.Errorf("no table exists for that data struct")
}
