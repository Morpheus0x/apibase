package db

import (
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/tables"
)

func GetUser(email string) (*tables.Users, *log.Error) {
	return nil, log.NewErrorWithType(log.ErrNotImplemented, "WIP")
}
