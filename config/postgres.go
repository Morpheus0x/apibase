package config

import (
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
)

func (ab *ApiBase) PostgresInit() *log.Error {
	var err *log.Error
	ab.ApiConfig.DB, err = db.PostgresInit(ab.Postgres, ab.BaseConfig)
	return err
}
