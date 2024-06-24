package sqlite

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Table struct {
	Name string
	Data interface{}

	raw_schema string // only used for table migration
}

func (t *Table) Schema() (string, error) {
	if t.raw_schema != "" {
		return fmt.Sprintf(t.raw_schema, t.Name), nil
	}
	return t.parseSchemaStruct()
}

func (t *Table) parseSchemaStruct() (string, error) {
	val := reflect.ValueOf(t.Data)
	if val.Kind() != reflect.Struct {
		return "", fmt.Errorf("input data must be a struct")
	}

	var columns, uniques []string
	primaryKey := ""
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)

		// skip protobuf internal struct fields
		if field.Name == "state" || field.Name == "sizeCache" || field.Name == "unknownFields" {
			continue
		}

		dbTag, ok := field.Tag.Lookup("db")
		if !ok || dbTag == "" {
			return "", fmt.Errorf("db tag value must be set for every database schema struct, table: %s, field: %s", t.Name, field.Name)
		}
		columnDefinition, colType := goTypeToSQLType(field.Type) // Golang to SQLite Type

		primaryTag, ok := field.Tag.Lookup("primary")
		if ok && primaryKey != "" {
			return "", fmt.Errorf("invalid primary key set, must be yes or auto and isn't allowed to be set twice per struct")
		}
		if ok && primaryTag != "" && primaryKey == "" {
			if primaryTag == "yes" {
				primaryKey = dbTag
				columnDefinition += " PRIMARY KEY"
			} else if primaryTag == "auto" {
				if primaryKey != "" {
					return "", fmt.Errorf("multiple primary keys defined")
				}
				primaryKey = dbTag
				columnDefinition = "INTEGER PRIMARY KEY AUTOINCREMENT"
			}
		}

		defaultTag, ok := field.Tag.Lookup("default")
		if ok && primaryTag != "" {
			return "", fmt.Errorf("primary and default can't be specified on the same column")
		}
		if ok {
			switch defaultTag {
			case "!null":
				columnDefinition += " NOT NULL"
			case "now":
				columnDefinition += " DEFAULT CURRENT_TIMESTAMP"
			case "":
				columnDefinition += " DEFAULT ''"
			default:
				defaultString, err := formatColumnDefault(defaultTag, colType)
				if err != nil {
					return "", err
				}
				columnDefinition += defaultString
			}
		}

		uniqueTag, ok := field.Tag.Lookup("unique")
		if ok {
			if uniqueTag == "" {
				uniques = append(uniques, fmt.Sprintf("\tUNIQUE(%s)", dbTag))
			} else {
				uniqueColumns := strings.Split(uniqueTag, ",")
				for i, col := range uniqueColumns {
					uniqueColumns[i] = strings.TrimSpace(col)
				}
				uniqueColumns = append([]string{dbTag}, uniqueColumns...) // Prepend dbTag for multi-column unique constraint
				uniques = append(uniques, fmt.Sprintf("\tUNIQUE(%s)", strings.Join(uniqueColumns, ", ")))
			}
		}

		newColumn := fmt.Sprintf("\t%s %s", dbTag, columnDefinition)
		columns = append(columns, newColumn)
	}

	if primaryKey == "" {
		return "", fmt.Errorf("no primary key defined")
	}
	if len(columns) < 1 {
		return "", fmt.Errorf("must be at least one column")
	}

	definitions := strings.Join(columns, ",\n")
	settings := ""
	if len(uniques) > 0 {
		settings = ",\n" + strings.Join(uniques, ",\n")
	}

	query := fmt.Sprintf("CREATE TABLE %s (\n%s%s\n)", t.Name, definitions, settings)
	return query, nil
}

// goTypeToSQLType converts a Go type to an SQL type.
// This function needs to be expanded based on your specific type mapping.
func goTypeToSQLType(t reflect.Type) (string, string) {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INTEGER", "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INTEGER", "int"
	case reflect.Float32, reflect.Float64:
		return "REAL", "float"
	case reflect.String:
		return "TEXT", "string"
	case reflect.Bool:
		return "BOOLEAN", "bool"
	case reflect.Struct: // assuming time.Time
		if t.Name() == "Time" && t.PkgPath() == "time" {
			return "DATETIME", "time"
		}
		return "TEXT", "interface"
	default:
		return "TEXT", "string"
	}
}

func formatColumnDefault(value string, colType string) (string, error) {
	switch colType {
	case "int":
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return "", fmt.Errorf("unable to parse default int value '%s', error: %v", value, err)
		}
		return fmt.Sprintf(" DEFAULT %d", parsed), nil
	case "float":
		parsed, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return "", fmt.Errorf("unable to parse default float value '%s', error: %v", value, err)
		}
		return fmt.Sprintf(" DEFAULT %f", parsed), nil
	case "string":
		return fmt.Sprintf(" DEFAULT '%s'", value), nil
	case "bool":
		if value == "true" || value == "1" {
			return " NOT NULL DEFAULT 1", nil
		} else if value == "false" || value == "0" {
			return " NOT NULL DEFAULT 0", nil
		} else {
			return "", fmt.Errorf("default value for bool is invalid: %s, must bee true, false, 1 or 0", value)
		}
	case "time":
		return "", fmt.Errorf("custom default time not yet supported")
	case "interface":
		return "", fmt.Errorf("unable to set default for unknown datatype")
	default:
		return "", fmt.Errorf("column default parsing, unknown datatype: %s", value)
	}
}
