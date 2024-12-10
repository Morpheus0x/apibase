package config

import "gopkg.cc/apibase/log"

var (
	ErrTomlParsing = log.RegisterErrType("toml parsing failed")
)
