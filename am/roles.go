package am

// Definition of roles
const (
	SystemRole        = "role:system"
	SystemSupportRole = "role:system_support"
	OwnerRole         = "role:owner"
	AdminRole         = "role:administrator"
	AuditorRole       = "role:auditor"
	EditorRole        = "role:editor"
	ReviewerRole      = "role:reviewer"
)

// RNSystem System Resource Name for allowing system/support access to all services
var RNSystem = "lrn:service:<.*>"

// Role represents roles to have policies applied to them
type Role struct {
	OrgID    int    `json:"org_id"`    // Organization ID
	RoleName string `json:"role_name"` // Friendly name of the role
	ID       string `json:"role_id"`   // ID is the role's unique id.
	Members  []int  `json:"member_id"` // Members who belong to the role.
}
