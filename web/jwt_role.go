package web

import "gopkg.cc/apibase/table"

// intentionally obfuscated json keys for security and bandwidth savings
type JwtRole struct {
	OrgView  bool `json:"a" toml:"org_view"`
	OrgEdit  bool `json:"b" toml:"org_edit"`
	OrgAdmin bool `json:"c" toml:"org_admin"`
}

type JwtRoles map[int]JwtRole

func JwtRolesFromTable(roles []table.UserRole) JwtRoles {
	jwtRoles := JwtRoles{}
	for _, r := range roles {
		jwtRoles[r.OrgID] = JwtRole{
			OrgView:  r.OrgView,
			OrgEdit:  r.OrgEdit,
			OrgAdmin: r.OrgAdmin,
		}
	}
	return jwtRoles
}

func (role JwtRole) GetTable(userId int, orgId int) table.UserRole {
	return table.UserRole{
		UserID:   userId,
		OrgID:    orgId,
		OrgView:  role.OrgView,
		OrgEdit:  role.OrgEdit,
		OrgAdmin: role.OrgAdmin,
	}
}
