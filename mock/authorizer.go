package mock

import "github.com/linkai-io/am/am"

type Authorizer struct {
	IsAllowedFn      func(subject, resource, action string) error
	IsAllowedInvoked bool

	IsUserAllowedFn      func(orgID, userID int, resource, action string) error
	IsUserAllowedInvoked bool

	GetRolesFn      func(orgID, userID int) ([]*am.Role, error)
	GetRolesInvoked bool
}

func (a *Authorizer) IsAllowed(subject, resource, action string) error {
	a.IsAllowedInvoked = true
	return a.IsAllowedFn(subject, resource, action)
}

func (a *Authorizer) IsUserAllowed(orgID, userID int, resource, action string) error {
	a.IsUserAllowedInvoked = true
	return a.IsUserAllowedFn(orgID, userID, resource, action)
}

func (a *Authorizer) GetRoles(orgID, userID int) ([]*am.Role, error) {
	a.GetRolesInvoked = true
	return a.GetRolesFn(orgID, userID)
}
