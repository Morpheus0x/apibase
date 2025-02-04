package table

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Source: https://stackoverflow.com/a/33678050

// Duration is used to convert between a bigint in Postgres and time.Duration
type Duration time.Duration

// Value converts Duration to a primitive value ready to written to a database.
func (d Duration) Value() (driver.Value, error) {
	return driver.Value(int64(d)), nil
}

// Scan reads a Duration value from database driver type.
func (d *Duration) Scan(raw interface{}) error {
	switch v := raw.(type) {
	case int64:
		*d = Duration(v)
	case nil:
		*d = Duration(0)
	default:
		return fmt.Errorf("cannot sql.Scan() table.Duration from: %#v", v)
	}
	return nil
}
