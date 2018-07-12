package am

import "context"

// Policy to be applied to a role via policy service/role service
type Policy struct {
	Subjects  []string
	Actions   []string
	Resources []string
}

// PolicyService is for managing policies that can be applied to roles
type PolicyService interface {
	AddPolicy(ctx context.Context, orgID, requesterUserID int, policy Policy) error    // creates a new policy
	UpdatePolicy(ctx context.Context, orgID, requesterUserID int, policy Policy) error // updates a policy
	NewOrgPolicies(ctx context.Context, orgID int) error                               // creates the initial set of policies for different groups
}
