package webtype

import "gopkg.cc/apibase/tables"

// intentionally obfuscated json keys for security and bandwidth savings
type JwtRole struct {
	OrgView  bool `json:"a"`
	OrgEdit  bool `json:"b"`
	OrgAdmin bool `json:"c"`
}

type JwtRoles map[int]JwtRole

func JwtRolesFromTable(roles []tables.UserRoles) JwtRoles {
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

// func (roles JwtRoles) GetTableUserRoles(userId int) []tables.UserRoles {
// 	var userRoles []tables.UserRoles
// 	for orgID, r := range roles {
// 		userRoles = append(userRoles, tables.UserRoles{
// 			UserID:   userId,
// 			OrgID:    orgID,
// 			OrgView:  r.OrgView,
// 			OrgEdit:  r.OrgEdit,
// 			OrgAdmin: r.OrgAdmin,
// 		})
// 	}
// 	return userRoles
// }
