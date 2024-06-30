package sqlite

import (
	"fmt"
	"reflect"
	"strings"
)

func cast[T any](input []interface{}) []*T {
	out := []*T{}
	for _, in := range input {
		out = append(out, (in).(*T))
	}
	return out
}

func SelectAll[T any](s *SQLite) ([]*T, error) {
	return Where[T](s, "")
}

// Query table derived from out with where clause passed as string
// return: changes out pointer to data returned from sqlite db
func Where[T any](s *SQLite, where string) ([]*T, error) {
	// TODO: check if first char of where is a space, if not, return error
	var myType [0]T
	table, err := s.getTable(reflect.TypeOf(myType).Elem())
	if err != nil {
		return []*T{}, err
	}
	columns, err := table.ColumnQueryString()
	if err != nil {
		return []*T{}, err
	}
	query := fmt.Sprintf("SELECT %s FROM %s%s;", columns, table.Name, where)
	outData, err := s.sqliteQuery(table, query)
	if err != nil {
		return []*T{}, err
	}
	return cast[T](outData), nil
}

// full SQL query without FROM ...
func Raw[T any](s *SQLite, query string) ([]*T, error) {
	var myType [0]T
	table, err := s.getTable(reflect.TypeOf(myType).Elem())
	if err != nil {
		return []*T{}, err
	}

	// Check if query matches outType
	nextIsTableName := false
	for _, s := range strings.Split(query, " ") {
		if nextIsTableName {
			if s != table.Name {
				return []*T{}, fmt.Errorf("table name in query doesn't match table from out datatype")
			}
			break
		}
		if s == "FROM" {
			nextIsTableName = true
		}
	}

	outData, err := s.sqliteQuery(table, query)
	if err != nil {
		return []*T{}, err
	}
	return cast[T](outData), err
}

// returns array of pointers
func (s *SQLite) sqliteQuery(table Table, query string) ([]interface{}, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return []interface{}{nil}, err
	}
	out, err := s.scanRows(rows, reflect.TypeOf(table.Data))
	return out, err
}

// TODO: shoud this return Table or string with table name?
func (s *SQLite) getTable(tableType reflect.Type) (Table, error) {
	for _, t := range s.tables {
		if reflect.TypeOf(t.Data) == tableType {
			return t, nil
		}
	}
	return Table{}, fmt.Errorf("no table exists for that data struct")
}
