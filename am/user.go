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
	OrgID                      int    `json:"org_id"`
	OrgCID                     string `json:"org_customer_id"`
	UserCID                    string `json:"user_customer_id"`
	UserID                     int    `json:"user_id"`
	UserEmail                  string `json:"user_email"`
	FirstName                  string `json:"first_name"`
	LastName                   string `json:"last_name"`
	StatusID                   int    `json:"status_id"`
	CreationTime               int64  `json:"creation_time"`
	Deleted                    bool   `json:"deleted"`
	AgreementAccepted          bool   `json:"agreement_accepted"`
	AgreementAcceptedTimestamp int64  `json:"agreement_accepted_timestamp"`
	LastLoginTimestamp         int64  `json:"last_login_timestamp"`
}

// UserContext interface for passing contextual data about a request for tracking & auth
type UserContext interface {
	GetTraceID() string
	GetOrgID() int
	GetOrgCID() string
	GetUserID() int
	GetUserCID() string
	GetRoles() []string
	GetIPAddress() string
	GetSubscriptionID() int32
}

// UserContextData for contextual information about a user
type UserContextData struct {
	TraceID        string   `json:"trace_id"`
	OrgID          int      `json:"org_id"`
	OrgCID         string   `json:"org_customer_id"`
	UserID         int      `json:"user_id"`
	UserCID        string   `json:"user_cid"`
	Roles          []string `json:"roles"`
	IPAddress      string   `json:"ip_address"`
	SubscriptionID int32    `json:"subscription_id"`
}

// NewUserContext creates user contextual data
func NewUserContext(orgID, userID int, orgCID, userCID, traceID, ipAddress string, roles []string, subscriptionID int32) *UserContextData {
	return &UserContextData{
		TraceID:        traceID,
		OrgID:          orgID,
		OrgCID:         orgCID,
		UserID:         userID,
		UserCID:        userCID,
		Roles:          roles,
		IPAddress:      ipAddress,
		SubscriptionID: subscriptionID,
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

// GetUserCID returns this context's user custom id
func (u *UserContextData) GetUserCID() string {
	return u.UserCID
}

// GetRoles returns this context's roles
func (u *UserContextData) GetRoles() []string {
	return u.Roles
}

// GetIPAddress returns this context's user ip address
func (u *UserContextData) GetIPAddress() string {
	return u.IPAddress
}

// GetSubscriptionID returns this context's user subscription level
func (u *UserContextData) GetSubscriptionID() int32 {
	return u.SubscriptionID
}

// UserFilter for limiting results from User List
type UserFilter struct {
	Start   int         `json:"start"`
	Limit   int         `json:"limit"`
	OrgID   int         `json:"org_id"`
	Filters *FilterType `json:"filters"`
}

// UserService for managing access to users
type UserService interface {
	Init(config []byte) error
	Get(ctx context.Context, userContext UserContext, userEmail string) (oid int, user *User, err error)
	GetWithOrgID(ctx context.Context, userContext UserContext, orgID int, userCID string) (oid int, user *User, err error)
	GetByID(ctx context.Context, userContext UserContext, userID int) (oid int, user *User, err error)
	GetByCID(ctx context.Context, userContext UserContext, userCID string) (oid int, user *User, err error)
	List(ctx context.Context, userContext UserContext, filter *UserFilter) (oid int, users []*User, err error)
	Create(ctx context.Context, userContext UserContext, user *User) (oid int, uid int, ucid string, err error)
	Update(ctx context.Context, userContext UserContext, user *User, userID int) (oid int, uid int, err error)
	Delete(ctx context.Context, userContext UserContext, userID int) (oid int, err error)
	AcceptAgreement(ctx context.Context, userContext UserContext, accepted bool) (oid int, uid int, err error)
}
