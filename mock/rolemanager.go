package mock

import "github.com/linkai-io/am/am"

type RoleManager struct {
	CreateRoleFn      func(*am.Role) (string, error)
	CreateRoleInvoked bool

	DeleteRoleFn      func(orgID int, roleID string) error
	DeleteRoleInvoked bool

	AddMembersFn      func(orgID int, roleID string, members []int) error
	AddMembersInvoked bool

	RemoveMembersFn      func(orgID int, roleID string, members []int) error
	RemoveMembersInvoked bool

	FindByMemberFn      func(orgID int, member int, limit, offset int) ([]*am.Role, error)
	FindByMemberInvoked bool

	GetFn      func(orgID int, roleID string) (*am.Role, error)
	GetInvoked bool

	GetByNameFn      func(orgID int, roleName string) (*am.Role, error)
	GetByNameInvoked bool

	ListFn      func(orgID int, limit, offset int) ([]*am.Role, error)
	ListInvoked bool
}

func (r *RoleManager) CreateRole(role *am.Role) (string, error) {
	r.CreateRoleInvoked = true
	return r.CreateRoleFn(role)
}

func (r *RoleManager) DeleteRole(orgID int, roleID string) error {
	r.DeleteRoleInvoked = true
	return r.DeleteRoleFn(orgID, roleID)
}

func (r *RoleManager) AddMembers(orgID int, roleID string, members []int) error {
	r.AddMembersInvoked = true
	return r.AddMembersFn(orgID, roleID, members)
}

func (r *RoleManager) RemoveMembers(orgID int, roleID string, members []int) error {
	r.RemoveMembersInvoked = true
	return r.RemoveMembersFn(orgID, roleID, members)
}

func (r *RoleManager) FindByMember(orgID int, member int, limit, offset int) ([]*am.Role, error) {
	r.FindByMemberInvoked = true
	return r.FindByMemberFn(orgID, member, limit, offset)
}

func (r *RoleManager) Get(orgID int, roleID string) (*am.Role, error) {
	r.GetInvoked = true
	return r.GetFn(orgID, roleID)
}

func (r *RoleManager) GetByName(orgID int, roleID string) (*am.Role, error) {
	r.GetByNameInvoked = true
	return r.GetByNameFn(orgID, roleID)
}

func (r *RoleManager) List(orgID int, limit, offset int) ([]*am.Role, error) {
	r.ListInvoked = true
	return r.ListFn(orgID, limit, offset)
}
