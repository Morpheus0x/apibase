package web

import "time"

const (
	TOKEN_ACCESS_VALIDITY  = time.Minute * 15
	TOKEN_REFRESH_VALIDITY = time.Hour * 24 * 30
)

var DefaultRole = JwtRole{OrgView: true, OrgEdit: false, OrgAdmin: false}
