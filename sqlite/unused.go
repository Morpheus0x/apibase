package sqlite

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

func (t Table) CreateOrUpdateRow(data interface{}) error {
	insertInsteadOfUpdate := false
	var idValue uint32 = 0
	// Ensure that data is a pointer to a struct
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("data must be a non-nil pointer to a struct")
	}

	// Dereference the pointer to get the struct
	structValue := val.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a pointer to a struct")
	}

	var columns []string
	var placeholders []string
	var updateColumns []string
	var values []interface{}

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		value := field.Interface()
		fieldType := structValue.Type().Field(i)

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
				continue
			}
			idValue = value.(uint32)
		}

		// Skip if field is time, has default tag set, and value is default time "zero"
		if structValue.Field(i).Type() == reflect.TypeOf(time.Time{}) {
			// if reflect.TypeOf(structValue.Field(i)) == time.Time {
			// if structValue.Type() == reflect.TypeOf(time.Time{}) {
			// G.LOG.Noticef("%s default empty value: %v, is zero? %v, default tag: '%s', insertInsteadOfUpdate: %v",
			// 	dbTag,
			// 	field.Interface(),
			// 	field.Interface().(time.Time).IsZero(),
			// 	fieldType.Tag.Get("default"),
			// 	insertInsteadOfUpdate,
			// )
			if fieldType.Tag.Get("default") != "" && value.(time.Time).IsZero() {
				continue
			} else {
				// Default time format contains +0100 timezone offset, this manual formatting disables that
				value = value.(time.Time).Format(G.FORMAT_SQLITE_DATETIME)
			}
		}

		columns = append(columns, dbTag)
		placeholders = append(placeholders, "?")
		updateColumns = append(updateColumns, fmt.Sprintf("%s = ?", dbTag))
		values = append(values, value)
	}

	var query string
	if insertInsteadOfUpdate {
		query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			t.Name, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	} else {
		// Docs: https://www.sqlitetutorial.net/sqlite-update/
		query = fmt.Sprintf("UPDATE %s SET %s WHERE id = %d",
			t.Name, strings.Join(updateColumns, ", "), idValue)
	}

	_, err := db.Exec(query, values...)
	if err != nil {
		return err
	}

	return nil
}
