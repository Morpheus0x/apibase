package sqlite

import (
	"fmt"
	"reflect"
	"strings"
)

func Cast[T any](input []interface{}, err error) ([]T, error) {
	out := []T{}
	if err != nil {
		return out, err
	}
	for _, in := range input {
		out = append(out, (in).(T))
	}
	return out, nil
}

// Query table derived from out with where clause passed as string
// return: changes out pointer to data returned from sqlite db
func (s *SQLite) Where(where string, outType any) ([]interface{}, error) {
	// TODO: check if first char of where is a space, if not, return error
	table, err := s.getTable(reflect.TypeOf(outType))
	if err != nil {
		return []interface{}{}, err
	}
	columns, err := table.ColumnQueryString()
	if err != nil {
		return []interface{}{}, err
	}
	query := fmt.Sprintf("SELECT %s FROM %s%s;", columns, table.Name, where)
	outData, err := s.sqliteQuery(table, query)
	return outData, err
}

// full SQL query without FROM ...
func (s *SQLite) Raw(query string, outType any) ([]interface{}, error) {
	table, err := s.getTable(reflect.TypeOf(outType))
	if err != nil {
		return []interface{}{}, err
	}

	// Check if query matches outType table
	nextIsTableName := false
	for _, s := range strings.Split(query, " ") {
		if nextIsTableName {
			if s != table.Name {
				return []interface{}{}, fmt.Errorf("table name in query doesn't match table from out datatype")
			}
			break
		}
		if s == "FROM" {
			nextIsTableName = true
		}
	}

	outData, err := s.sqliteQuery(table, query)
	return outData, err
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
func (s *SQLite) getTable(tableType reflect.Type) (Table, error) {
	for _, t := range s.tables {
		if reflect.TypeOf(t.Data) == tableType {
			return t, nil
		}
	}
	return Table{}, fmt.Errorf("no table exists for that data struct")
}
