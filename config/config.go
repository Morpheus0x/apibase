package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"gopkg.cc/apibase/db"
	"gopkg.cc/apibase/log"
)

type ApiBase struct {
	Postgres db.PostgresConfig `toml:"postgres"`
}

func LoadToml(path string) (*ApiBase, *log.Error) {
	_, err := os.Stat(path)
	if err != nil {
		return &ApiBase{}, log.NewErrorWithTypef(ErrTomlParsing, "unable to read toml file: %s", err.Error())
	}
	var apiBase ApiBase
	_, err = toml.DecodeFile(path, &apiBase)
	if err != nil {
		return &ApiBase{}, log.NewErrorWithTypef(ErrTomlParsing, "unable to parse toml: %s", err.Error())
	}
	return &apiBase, log.ErrorNil()
}
