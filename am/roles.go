package am

// DefaultRole created on organization setup
type DefaultRole int32

// Definition of roles
const (
	Owner DefaultRole = iota << 1
	Administrator
	Auditor
	Editor
	Reviewer
)

// Role represents roles to have policies applied to them
type Role struct {
	OrgID   int32   `json:"org_id"`    // Organization ID
	ID      string  `json:"role_id"`   // ID is the role's unique id.
	Members []int32 `json:"member_id"` // Members who belong to the role.
}

// RoleMap for string definitions
var RoleMap = map[DefaultRole]string{
	Owner:         "Owner",
	Administrator: "Administrator",
	Auditor:       "Auditor",
	Editor:        "Editor",
	Reviewer:      "Reviewer",
}
