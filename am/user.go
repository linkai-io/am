package am

import "context"

const (
	// RNUserSystem system only access
	RNUserSystem = "lrn:service:user:feature:system"
	// RNUserManage organization specific management
	RNUserManage   = "lrn:service:user:feature:manage"
	RNUserSelf     = "lrn:service:user:feature:self"
	UserServiceKey = "userservice"
)

const (
	UserStatusDisabled        = 1
	UserStatusAwaitActivation = 100
	UserStatusActive          = 1000
	UserStatusSystem          = 9999
)

// User represents a user of an organization that has subscribed to our service
type User struct {
	OrgID        int    `json:"org_id"`
	OrgCID       string `json:"org_customer_id"`
	UserCID      string `json:"user_customer_id"`
	UserID       int    `json:"user_id"`
	UserEmail    string `json:"user_email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	StatusID     int    `json:"status_id"`
	CreationTime int64  `json:"creation_time"`
	Deleted      bool   `json:"deleted"`
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
	Start             int   `json:"start"`
	Limit             int   `json:"limit"`
	OrgID             int   `json:"org_id"`
	SinceCreationTime int64 `json:"since_creation_time"`
	WithStatus        bool  `json:"with_status"`
	StatusValue       int   `json:"status_value"`
	WithDeleted       bool  `json:"with_deleted"`
	DeletedValue      bool  `json:"deleted_value"`
}

// UserService for managing access to users
type UserService interface {
	Init(config []byte) error
	Get(ctx context.Context, userContext UserContext, userEmail string) (oid int, user *User, err error)
	GetWithOrgID(ctx context.Context, userContext UserContext, orgID int, userEmail string) (oid int, user *User, err error)
	GetByID(ctx context.Context, userContext UserContext, userID int) (oid int, user *User, err error)
	GetByCID(ctx context.Context, userContext UserContext, userCID string) (oid int, user *User, err error)
	List(ctx context.Context, userContext UserContext, filter *UserFilter) (oid int, users []*User, err error)
	Create(ctx context.Context, userContext UserContext, user *User) (oid int, uid int, ucid string, err error)
	Update(ctx context.Context, userContext UserContext, user *User, userID int) (oid int, uid int, err error)
	Delete(ctx context.Context, userContext UserContext, userID int) (oid int, err error)
}
