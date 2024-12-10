package sqlite

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Co-Written with ChatGPT

// return array of pointers
func (s *SQLite) scanRows(rows *sql.Rows, targetType reflect.Type) ([]interface{}, error) {
	var outArray []interface{}
	if targetType.Kind() != reflect.Struct {
		return outArray, fmt.Errorf("targetType must be a struct")
	}
	validTargetFieldCount := 0
	targetFieldNames := []string{}
	targetFields := make([]int, targetType.NumField())
	targetFieldTypes := map[int]interface{}{}
	var pbTime *timestamppb.Timestamp
	var pbTimeAddr **timestamppb.Timestamp
	var goTimeStr string
	for i := 0; i < targetType.NumField(); i++ {
		dbField := targetType.Field(i).Tag.Get("db")
		if dbField == "" {
			continue
		}
		mappedType := reflect.New(targetType).Elem().Field(i).Addr().Interface()
		if reflect.TypeOf(mappedType) == reflect.TypeOf(pbTime) || reflect.TypeOf(mappedType) == reflect.TypeOf(pbTimeAddr) {
			mappedType = reflect.New(reflect.TypeOf(goTimeStr)).Elem().Addr().Interface()
		}
		targetFieldTypes[validTargetFieldCount] = mappedType
		targetFields[validTargetFieldCount] = i
		targetFieldNames = append(targetFieldNames, dbField)
		validTargetFieldCount++
	}
	columns, err := rows.Columns()
	if err != nil {
		return outArray, err
	}
	if len(columns) > validTargetFieldCount {
		return outArray, fmt.Errorf("targetType has less fields than required by query result")
	}
	for i, c := range columns {
		if targetFieldNames[i] != c {
			return outArray, fmt.Errorf("targetType next field '%s' doesn't match next column '%s'", targetFieldNames[i], c)
		}
	}

	for rows.Next() {
		fields := make([]interface{}, len(columns))
		for i := 0; i < len(columns); i++ {
			fields[i] = targetFieldTypes[i]
		}

		if err := rows.Scan(fields...); err != nil {
			return outArray, err
		}

		outRow := reflect.New(targetType).Elem()
		for i := 0; i < len(fields); i++ {
			dbField := targetType.Field(i).Tag.Get("db")
			var val reflect.Value
			targetTypeToParse := reflect.TypeOf(reflect.New(targetType).Elem().Field(targetFields[i]).Addr().Elem().Interface())
			if targetTypeToParse == reflect.TypeOf(pbTime) || targetTypeToParse == reflect.TypeOf(pbTimeAddr) {
				timeString := reflect.ValueOf(fields[i]).Elem().Interface().(string)
				time, err := time.Parse(s.config.SQLITE_DATETIME_FORMAT, timeString)
				if err != nil {
					return outArray, fmt.Errorf("unable to parse time (%s) for column %s", timeString, dbField)
				}
				pbNewTime := timestamppb.New(time)
				val = reflect.ValueOf(pbNewTime)
			} else {
				val = reflect.ValueOf(fields[i]).Elem() // Elem() to dereference pointer in fields[i]
			}
			outRow.Field(targetFields[i]).Set(val)
		}

		outArray = append(outArray, outRow.Addr().Interface())
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return outArray, nil
}
