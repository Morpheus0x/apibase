package base

import "gopkg.cc/apibase/log"

var (
	ErrTomlParsing    = log.RegisterErrType("toml parsing failed")
	ErrApiBaseCleanup = log.RegisterErrType("issue with Apibase cleanup")
)
