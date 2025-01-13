package web

import "gopkg.cc/apibase/errx"

var (
	ErrTokenValidate = errx.NewType("token validation")
)
