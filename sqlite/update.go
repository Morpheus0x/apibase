package sqlite

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Returns Inserted ID and possibly error
func CreateOrUpdateRow(s *SQLite, data interface{}) (int64, error) {
	insertInsteadOfUpdate := false
	var idValue uint32 = 0
	// Ensure that data is a pointer to a struct
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return -1, fmt.Errorf("data must be a non-nil pointer to a struct")
	}

	// Dereference the pointer to get the struct
	structValue := val.Elem()
	if structValue.Kind() != reflect.Struct {
		return -1, fmt.Errorf("data must be a pointer to a struct")
	}

	table, err := s.getTable(reflect.TypeOf(data).Elem())
	if err != nil {
		return -1, fmt.Errorf("data must be a known table struct")
	}

	skipPrivateFields := make([]bool, structValue.NumField())
	visibleFields := reflect.VisibleFields(structValue.Type())
	if structValue.NumField() != len(visibleFields) {
		return -1, fmt.Errorf("length of struct and visible fields of struct don't match")
	}
	for c, f := range visibleFields {
		skipPrivateFields[c] = !f.IsExported()
	}

	var columns []string
	var placeholders []string
	var updateColumns []string
	var values []interface{}

	for i := 0; i < structValue.NumField(); i++ {
		if skipPrivateFields[i] {
			continue
		}
		field := structValue.Field(i)
		value := field.Interface()
		fieldType := structValue.Type().Field(i)
		// fmt.Printf("field i: %d, value: %v\n", i, value)

		// Get the db tag
		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" {
			continue
		}

		// Skip zero fields to avoid the zero Value issue
		if !field.IsValid() || (field.Kind() == reflect.Ptr && field.IsNil()) {
			continue
		}

		// Skip id field, if kept at default 0, this should create new row
		if dbTag == "id" {
			if value.(uint32) == 0 {
				insertInsteadOfUpdate = true
				// fmt.Print("insert instead of update\n")
				continue
			}
			idValue = value.(uint32)
		}

		// Skip if field is time, has default tag set, and value is default time "zero"
		if structValue.Field(i).Type() == reflect.TypeOf(&timestamppb.Timestamp{}) {
			// fmt.Printf("TIME!\n")
			// if reflect.TypeOf(structValue.Field(i)) == time.Time {
			// if structValue.Type() == reflect.TypeOf(time.Time{}) {
			// G.LOG.Noticef("%s default empty value: %v, is zero? %v, default tag: '%s', insertInsteadOfUpdate: %v",
			// 	dbTag,
			// 	field.Interface(),
			// 	field.Interface().(time.Time).IsZero(),
			// 	fieldType.Tag.Get("default"),
			// 	insertInsteadOfUpdate,
			// )
			time := value.(*timestamppb.Timestamp).AsTime()

			if fieldType.Tag.Get("default") != "" && time.IsZero() {
				continue
			} else {
				// Default time format contains +0100 timezone offset, this manual formatting disables that
				value = time.Format(s.config.FORMAT_SQLITE_DATETIME)
			}
		}
		// fmt.Printf("value parsed: %v, type: %v\n", value, structValue.Field(i).Type())

		columns = append(columns, dbTag)
		placeholders = append(placeholders, "?")
		updateColumns = append(updateColumns, fmt.Sprintf("%s = ?", dbTag))
		values = append(values, value)
	}

	var query string
	if insertInsteadOfUpdate {
		query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			table.Name, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	} else {
		// Docs: https://www.sqlitetutorial.net/sqlite-update/
		query = fmt.Sprintf("UPDATE %s SET %s WHERE id = %d",
			table.Name, strings.Join(updateColumns, ", "), idValue)
	}

	res, err := s.DB.Exec(query, values...)
	if err != nil {
		return -1, err
	}

	if insertInsteadOfUpdate {
		return res.LastInsertId()
	} else {
		return int64(idValue), nil
	}
}
