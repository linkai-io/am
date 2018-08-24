package auth

import "github.com/linkai-io/am/am"

// RoleManager interface for managing roles and members of roles/groups
type RoleManager interface {
	CreateRole(*am.Role) (string, error)
	DeleteRole(orgID int, roleID string) error
	AddMembers(orgID int, roleID string, members []int) error
	RemoveMembers(orgID int, roleID string, members []int) error
	FindByMember(orgID int, member int, limit, offset int) ([]*am.Role, error)
	Get(orgID int, roleID string) (*am.Role, error)
	GetByName(orgID int, roleName string) (*am.Role, error)
	List(orgID int, limit, offset int) ([]*am.Role, error)
}
