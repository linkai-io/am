package auth

import "gopkg.linkai.io/v1/repos/am/am"

// Authorizer interfaces between roles and policies to determine if a user is a member of a role allowed to access a resource
// note it does not determine authorization of organization data, that is done at the datastore access level.
type Authorizer interface {
	IsAllowed(subject, resource, action string) error
	IsUserAllowed(orgID, userID int32, resource, action string) error
	GetRoles(orgID, userID int32) ([]*am.Role, error)
}
