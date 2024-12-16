package config

import (
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
	"gopkg.cc/apibase/webtype"
)

type ApiBase struct {
	Postgres   db.PostgresConfig `toml:"postgres"`
	SQLite     db.SQLiteConfig   `toml:"sqlite"`
	BaseConfig db.BaseConfig     `toml:"baseconfig"`
	ApiConfig  webtype.ApiConfig `toml:"apiconfig"`
}

func LoadToml(path string) (*ApiBase, *log.Error) {
	apiBase := &ApiBase{}

	_, err := os.Stat(path)
	if err != nil {
		return apiBase, log.NewErrorWithTypef(ErrTomlParsing, "unable to read toml file: %s", err.Error())
	}
	_, err = toml.DecodeFile(path, apiBase)
	if err != nil {
		return apiBase, log.NewErrorWithTypef(ErrTomlParsing, "unable to parse toml: %s", err.Error())
	}

	apiBase.AddMissingDefaults()

	return apiBase, log.ErrorNil()
}

func (apiBase *ApiBase) AddMissingDefaults() {
	if reflect.ValueOf(apiBase.BaseConfig.DB_MAX_RECONNECT_ATTEMPTS).IsZero() {
		apiBase.BaseConfig.DB_MAX_RECONNECT_ATTEMPTS = DB_MAX_RECONNECT_ATTEMPTS
	}
	if reflect.ValueOf(apiBase.BaseConfig.DB_RECONNECT_TIMEOUT_SEC).IsZero() {
		apiBase.BaseConfig.DB_RECONNECT_TIMEOUT_SEC = DB_RECONNECT_TIMEOUT_SEC
	}
	if reflect.ValueOf(apiBase.BaseConfig.SQLITE_DATETIME_FORMAT).IsZero() {
		apiBase.BaseConfig.SQLITE_DATETIME_FORMAT = SQLITE_DATETIME_FORMAT
	}
}
