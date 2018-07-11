package am

// User represents a user of an organization that has subscribed to our service
type User struct {
	OrgID     int32  `json:"org_id"`
	UserID    int32  `json:"user_id"`
	UserEmail string `json:"user_email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UserContext interface for passing contextual data about a request for tracking & auth
type UserContext interface {
	GetTraceID() string
	GetOrgID() int32
	GetUserID() int32
	GetRoles() []string
	GetIPAddress() string
}

// UserContextData for contextual information about a user
type UserContextData struct {
	TraceID   string   `json:"trace_id"`
	OrgID     int32    `json:"org_id"`
	UserID    int32    `json:"user_id"`
	Roles     []string `json:"roles"`
	IPAddress string   `json:"ip_address"`
}

// NewUserContext creates user contextual data
func NewUserContext(orgID, userID int32, traceID, ipAddress string, roles []string) *UserContextData {
	return &UserContextData{
		TraceID:   traceID,
		OrgID:     orgID,
		UserID:    userID,
		Roles:     roles,
		IPAddress: ipAddress,
	}
}

// GetTraceID returns the id used for tracking requests
func (u *UserContextData) GetTraceID() string {
	return u.TraceID
}

// GetOrgID returns this context's org id
func (u *UserContextData) GetOrgID() int32 {
	return u.OrgID
}

// GetUserID returns this context's user id
func (u *UserContextData) GetUserID() int32 {
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
