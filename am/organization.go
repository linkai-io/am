package am

import "context"

const (
	// RNOrganizationSystem system only access (create/delete)
	RNOrganizationSystem = "lrn:service:organization:feature:system"
	// RNOrganizationManage organization specific management
	RNOrganizationManage   = "lrn:service:organization:feature:manage"
	OrganizationServiceKey = "orgservice"
)

const (
	OrgStatusDisabledPendingPayment = 1
	OrgStatusDisabledClosed         = 2
	OrgStatusDisabledLocked         = 3
	OrgStatusAwaitActivation        = 100
	OrgStatusActive                 = 1000

	SubscriptionPending       = 1
	SubscriptionOneTime       = 10
	SubscriptionMonthly       = 100
	SubscriptionMonthlySmall  = 101
	SubscriptionMonthlyMedium = 102
	SubscriptionEnterprise    = 1000
	SubscriptionSystem        = 9999
)

// Organization represents an organization that has subscribed to our service
type Organization struct {
	OrgID                      int    `json:"org_id"`
	OrgCID                     string `json:"org_customer_id"`
	OrgName                    string `json:"org_name"`
	OwnerEmail                 string `json:"owner_email"`
	UserPoolID                 string `json:"user_pool_id"`
	UserPoolAppClientID        string `json:"user_pool_app_client_id"`
	UserPoolAppClientSecret    string `json:"user_pool_app_client_secret"`
	IdentityPoolID             string `json:"identity_pool_id"`
	UserPoolJWK                string `json:"user_pool_jwk"`
	FirstName                  string `json:"first_name"`
	LastName                   string `json:"last_name"`
	Phone                      string `json:"phone"`
	Country                    string `json:"country"`
	StatePrefecture            string `json:"state_prefecture"`
	Street                     string `json:"street"`
	Address1                   string `json:"address1"`
	Address2                   string `json:"address2"`
	City                       string `json:"city"`
	PostalCode                 string `json:"postal_code"`
	CreationTime               int64  `json:"creation_time"`
	StatusID                   int    `json:"status_id"`
	Deleted                    bool   `json:"deleted"`
	SubscriptionID             int32  `json:"subscription_id"`
	LimitTLD                   int32  `json:"limit_tld"`
	LimitTLDReached            bool   `json:"limit_tld_reached"`
	LimitHosts                 int32  `json:"limit_hosts"`
	LimitHostsReached          bool   `json:"limit_hosts_reached"`
	LimitCustomWebFlows        int32  `json:"limit_custom_web_flows"`
	LimitCustomWebFlowsReached bool   `json:"limit_custom_web_flows_reached"`
	PortScanEnabled            bool   `json:"port_scan_enabled"`
	PaymentRequiredTimestamp   int64  `json:"payment_required_timestamp"`
	BillingPlanType            string `json:"billing_plan_type"`
	BillingPlanID              string `json:"billing_plan_id"`
	BillingSubscriptionID      string `json:"billing_subscription_id"`
	IsBetaPlan                 bool   `json:"is_beta_plan"`
}

// OrgFilter for filtering organization list results
type OrgFilter struct {
	Start   int         `json:"start"`
	Limit   int         `json:"limit"`
	Filters *FilterType `json:"filters"`
}

// OrganizationService manages access to organizations
type OrganizationService interface {
	Init(config []byte) error
	Get(ctx context.Context, userContext UserContext, orgName string) (oid int, org *Organization, err error)
	GetByCID(ctx context.Context, userContext UserContext, orgCID string) (oid int, org *Organization, err error)
	GetByID(ctx context.Context, userContext UserContext, orgID int) (oid int, org *Organization, err error)
	GetByAppClientID(ctx context.Context, userContext UserContext, orgAppClientID string) (oid int, org *Organization, err error)
	List(ctx context.Context, userContext UserContext, filter *OrgFilter) (orgs []*Organization, err error)
	Create(ctx context.Context, userContext UserContext, org *Organization, userCID string) (oid int, uid int, ocid string, ucid string, err error)
	Update(ctx context.Context, userContext UserContext, org *Organization) (oid int, err error)
	Delete(ctx context.Context, userContext UserContext, orgID int) (oid int, err error)
}
