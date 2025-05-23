package base

import "gopkg.cc/apibase/errx"

var (
	ErrTomlParsing    = errx.NewType("toml parsing failed")
	ErrApiBaseCleanup = errx.NewType("issue with Apibase cleanup")
	ErrApiRootParsing = errx.NewType("invalid api root cli arg")
	ErrEmailParsing   = errx.NewType("unable to parse email config")
)
