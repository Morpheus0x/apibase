package web

import "time"

const (
	TOKEN_ACCESS_VALIDITY      = time.Minute * 15
	TOKEN_REFRESH_VALIDITY     = time.Hour * 24 * 30
	TOKEN_COOKIE_EXPIRY_MARGIN = "20%"
	REFERRER_MAX_LENGTH        = 2048
)

var DefaultRole = JwtRole{OrgView: true, OrgEdit: false, OrgAdmin: false}

const GatewayTimeoutHTML = `<html>
	<head><title>502 Bad Gateway</title></head>
	<body>
		<center><h1>502 Bad Gateway</h1></center>
		<hr><center></center>
	</body>
</html>`
