package web

import "gopkg.cc/apibase/errx"

var (
	ErrTokenParsing  = errx.NewType("unable to parse token")
	ErrTokenValidate = errx.NewType("token is invalid")
)
