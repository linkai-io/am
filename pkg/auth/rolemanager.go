package auth

import "gopkg.linkai.io/v1/repos/am/am"

// RoleManager interface for managing roles and members of roles/groups
type RoleManager interface {
	CreateRole(*am.Role) (string, error)
	DeleteRole(orgID int32, roleID string) error
	AddMembers(orgID int32, roleID string, members []int32) error
	RemoveMembers(orgID int32, roleID string, members []int32) error
	FindByMember(orgID int32, member int32, limit, offset int) ([]*am.Role, error)
	Get(orgID int32, roleID string) (*am.Role, error)
	GetByName(orgID int32, roleName string) (*am.Role, error)
	List(orgID int32, limit, offset int) ([]*am.Role, error)
}
