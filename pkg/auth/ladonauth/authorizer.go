package ladonauth

import (
	"errors"
	"math"

	"github.com/ory/ladon"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
)

var (
	ErrNoRoleDefined = errors.New("unable to find role(s) for user")
)

// LadonAuthorizer authorizers that a role is allowed access to a resource
type LadonAuthorizer struct {
	manager     ladon.Manager
	warden      ladon.Warden
	roleManager auth.RoleManager
}

// NewLadonAuthorizer returns a new authorizer backed by the policy and role managers
func NewLadonAuthorizer(policyManager ladon.Manager, roleManager auth.RoleManager) *LadonAuthorizer {
	return &LadonAuthorizer{
		manager:     policyManager,
		warden:      &ladon.Ladon{Manager: policyManager},
		roleManager: roleManager,
	}
}

// IsAllowed checks that the subject is allowed to do action on resource, returns nil
// on is allowed, error otherwise.
func (a *LadonAuthorizer) IsAllowed(subject, resource, action string) error {
	request := &ladon.Request{
		Subject:  subject,
		Resource: resource,
		Action:   action,
	}
	return a.warden.IsAllowed(request)
}

// IsUserAllowed iterates over all roles this user has applied to them and checks that their
// role is allowed to acces the resource. If isAllowed never returns nil, that means at least
// one role is allowed access. Otherwise return the last error seen.
func (a *LadonAuthorizer) IsUserAllowed(orgID, userID int, resource, action string) error {
	var err error
	memberRoles, err := a.GetRoles(orgID, userID)
	if err != nil {
		return err
	}

	if len(memberRoles) == 0 {
		return ErrNoRoleDefined
	}

	for _, role := range memberRoles {
		err = a.IsAllowed(role.RoleName, resource, action)
		if err == nil {
			return nil
		}
	}
	return err
}

// GetRoles looks up all roles applied to the userID of orgID
func (a *LadonAuthorizer) GetRoles(orgID, userID int) ([]*am.Role, error) {
	return a.roleManager.FindByMember(orgID, userID, math.MaxInt32, 0)
}
