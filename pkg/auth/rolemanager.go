package auth

import "gopkg.linkai.io/v1/repos/am/am"

// RoleManager interface for managing roles and members of roles/groups
type RoleManager interface {
	Create(*am.Role) error
	Get(id string) (*am.Role, error)
	Delete(id string) error

	AddMembers(group string, members []string) error
	RemoveMembers(group string, members []string) error

	FindByMember(member string, limit, offset int) ([]am.Role, error)
	List(limit, offset int) ([]am.Role, error)
}
