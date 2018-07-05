package am

import "context"

type PolicyEffect int

const (
	Allow PolicyEffect = 1
	Deny  PolicyEffect = 2
)

type Policy struct {
	Subjects []string
	Actions  []string
	Effect   PolicyEffect
}

type PolicyService interface {
	AddPolicy(ctx context.Context, orgID, requesterUserID int32, policy Policy) error    // creates a new policy
	UpdatePolicy(ctx context.Context, orgID, requesterUserID int32, policy Policy) error // updates a policy
	NewOrgPolicies(ctx context.Context, orgID int32) error                               // creates the initial set of policies for different groups
}
