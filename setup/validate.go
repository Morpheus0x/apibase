package setup

import (
	"fmt"

	"gopkg.cc/apibase/db"
)

func validateDB(database db.DB) error {
	switch database.Kind {
	case db.SQLite:
		if database.SQLite == nil {
			return fmt.Errorf("no valid SQLite database adapter")
		}
	case db.PostgreSQL:
		if database.Gorm == nil {
			return fmt.Errorf("no valid PostgreSQL database adapter")
		}
	default:
		return fmt.Errorf("no valid DB specified")
	}
	return nil
}
