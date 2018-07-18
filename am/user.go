package am

import "context"

// User represents a user of an organization that has subscribed to our service
type User struct {
	OrgID     int    `json:"org_id"`
	OrgCID    string `json:"org_customer_id"`
	UserCID   string `json:"user_customer_id"`
	UserID    int    `json:"user_id"`
	UserEmail string `json:"user_email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UserContext interface for passing contextual data about a request for tracking & auth
type UserContext interface {
	GetTraceID() string
	GetOrgID() int
	GetUserID() int
	GetRoles() []string
	GetIPAddress() string
}

// UserContextData for contextual information about a user
type UserContextData struct {
	TraceID   string   `json:"trace_id"`
	OrgID     int      `json:"org_id"`
	OrgCID    string   `json:"org_customer_id"`
	UserID    int      `json:"user_id"`
	Roles     []string `json:"roles"`
	IPAddress string   `json:"ip_address"`
}

// NewUserContext creates user contextual data
func NewUserContext(orgID, userID int, orgCID, traceID, ipAddress string, roles []string) *UserContextData {
	return &UserContextData{
		TraceID:   traceID,
		OrgID:     orgID,
		OrgCID:    orgCID,
		UserID:    userID,
		Roles:     roles,
		IPAddress: ipAddress,
	}
}

// GetTraceID returns the id used for tracking requests
func (u *UserContextData) GetTraceID() string {
	return u.TraceID
}

// GetOrgCID returns this context's org customer id (facing)
func (u *UserContextData) GetOrgCID() string {
	return u.OrgCID
}

// GetOrgID returns this context's org id
func (u *UserContextData) GetOrgID() int {
	return u.OrgID
}

// GetUserID returns this context's user id
func (u *UserContextData) GetUserID() int {
	return u.UserID
}

// GetRoles returns this context's roles
func (u *UserContextData) GetRoles() []string {
	return u.Roles
}

// GetIPAddress returns this context's user ip address
func (u *UserContextData) GetIPAddress() string {
	return u.IPAddress
}

// UserFilter for limiting results from User List
type UserFilter struct {
	Start int
	Limit int
}

// UserService for managing access to users
type UserService interface {
	Get(ctx context.Context, userContext UserContext, userID int) (*User, error)
	GetByCUID(ctx context.Context, userContext UserContext, userCID string) (*User, error)
	List(ctx context.Context, userContext UserContext, filter *UserFilter) ([]*User, error)
	Delete(ctx context.Context, userContext UserContext, userID int) error
	Create(ctx context.Context, userContext UserContext, user *User) (string, error)
	Update(ctx context.Context, userContext UserContext, user *User) error
}
