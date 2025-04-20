package roles

import "github.com/google/uuid"

type OrgRole int8

const (
	NONE     OrgRole = -1
	READER   OrgRole = 0
	APPROVER OrgRole = 1
	OPERATOR OrgRole = 2
	MANAGER  OrgRole = 3
	ADMIN    OrgRole = 4
)

type OrgRoles map[uuid.UUID]OrgRole

func (o OrgRoles) GetRole(orgid uuid.UUID) OrgRole {
	if role, ok := o[orgid]; ok {
		return role
	}
	return NONE
}