package organization

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx"
	uuid "github.com/satori/go.uuid"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/auth"
)

// Service for interfacing with postgresql/rds
type Service struct {
	pool       *pgx.ConnPool
	config     *pgx.ConnPoolConfig
	authorizer auth.Authorizer
}

// New returns an empty Service
func New(authorizer auth.Authorizer) *Service {
	return &Service{authorizer: authorizer}
}

// Init by parsing the config and initializing the database pool
func (s *Service) Init(config []byte) error {
	var err error

	s.config, err = s.parseConfig(config)
	if err != nil {
		return err
	}

	if s.pool, err = pgx.NewConnPool(*s.config); err != nil {
		return err
	}

	return nil
}

// parseConfig parses the configuration options and validates they are sane.
func (s *Service) parseConfig(config []byte) (*pgx.ConnPoolConfig, error) {
	dbstring := string(config)
	if dbstring == "" {
		return nil, am.ErrEmptyDBConfig
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		return nil, am.ErrInvalidDBString
	}

	return &pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 50,
		AfterConnect:   s.afterConnect,
	}, nil
}

// afterConnect will iterate over prepared statements with keywords
func (s *Service) afterConnect(conn *pgx.Conn) error {
	for k, v := range queryMap {
		if _, err := conn.Prepare(k, v); err != nil {
			log.Printf("%s\n", k)
			return err
		}
	}
	return nil
}

// IsAuthorized checks if an action is allowed by a particular user
func (s *Service) IsAuthorized(ctx context.Context, userContext am.UserContext, resource, action string) bool {
	if err := s.authorizer.IsUserAllowed(userContext.GetOrgID(), userContext.GetUserID(), resource, action); err != nil {
		return false
	}
	return true
}

// Get organization by organization name, system user only.
func (s *Service) Get(ctx context.Context, userContext am.UserContext, orgName string) (*am.Organization, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "read") {
		return nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("orgByName", orgName))
}

// GetByCID organization by organization customer id
func (s *Service) GetByCID(ctx context.Context, userContext am.UserContext, orgCID string) (*am.Organization, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationManage, "read") {
		return nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("orgByCID", orgCID))
}

// GetByID organization by ID, system user only.
func (s *Service) GetByID(ctx context.Context, userContext am.UserContext, orgID int) (*am.Organization, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "read") {
		return nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("orgByID", orgID))
}

// get executes the scan against the previously created queryrow
func (s *Service) get(ctx context.Context, userContext am.UserContext, row *pgx.Row) (*am.Organization, error) {
	org := &am.Organization{}
	err := row.Scan(&org.OrgID, &org.OrgName, &org.OrgCID, &org.UserPoolID, &org.IdentityPoolID, &org.OwnerEmail, &org.FirstName, &org.LastName, &org.Phone,
		&org.Country, &org.StatePrefecture, &org.Street, &org.Address1, &org.Address2, &org.City, &org.PostalCode, &org.CreationTime, &org.StatusID, &org.Deleted, &org.SubscriptionID)
	return org, err
}

// List all organizations that match the supplied filter, system users only.
func (s *Service) List(ctx context.Context, userContext am.UserContext, filter *am.OrgFilter) ([]*am.Organization, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "read") {
		return nil, am.ErrUserNotAuthorized
	}
	organizations := make([]*am.Organization, 0)

	rows, err := s.pool.Query("orgList", filter.Start, filter.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		org := &am.Organization{}

		if err := rows.Scan(&org.OrgID, &org.OrgName, &org.OrgCID, &org.UserPoolID, &org.IdentityPoolID, &org.OwnerEmail, &org.FirstName, &org.LastName, &org.Phone,
			&org.Country, &org.StatePrefecture, &org.Street, &org.Address1, &org.Address2, &org.City, &org.PostalCode, &org.CreationTime, &org.StatusID, &org.Deleted, &org.SubscriptionID); err != nil {
			return nil, err
		}

		organizations = append(organizations, org)
	}

	return organizations, nil
}

// Create a new organization, and intialize the user + roles, system users only.
func (s *Service) Create(ctx context.Context, userContext am.UserContext, org *am.Organization) (orgCID string, userCID string, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "create") {
		return "", "", am.ErrUserNotAuthorized
	}
	tx, err := s.pool.Begin()
	defer tx.Rollback()

	oid := 0
	name := ""
	cid := ""
	err = tx.QueryRow("orgExists", org.OrgName, -1, "").Scan(&oid, &name, &cid)
	if err != nil && err != pgx.ErrNoRows {
		return "", "", err
	}

	if oid != 0 {
		return "", "", am.ErrOrganizationExists
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", "", err
	}
	orgCID = id.String()

	id, err = uuid.NewV4()
	if err != nil {
		return "", "", err
	}
	userCID = id.String()

	now := time.Now().UnixNano()
	if _, err = tx.Exec("orgCreate", org.OrgName, orgCID, org.UserPoolID, org.IdentityPoolID, org.OwnerEmail,
		org.FirstName, org.LastName, org.Phone, org.Country, org.StatePrefecture, org.Street, org.Address1,
		org.Address2, org.City, org.PostalCode, now, org.StatusID, org.SubscriptionID, userCID, org.OwnerEmail,
		org.FirstName, org.LastName); err != nil {

		return "", "", err
	}

	err = tx.Commit()
	return orgCID, userCID, err
}

// Update allows the customer to update the details of their organization
func (s *Service) Update(ctx context.Context, userContext am.UserContext, org *am.Organization) error {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationManage, "update") {
		return am.ErrUserNotAuthorized
	}
	return nil
}

// Delete the organization
func (s *Service) Delete(ctx context.Context, userContext am.UserContext, orgID int) error {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "delete") {
		return am.ErrUserNotAuthorized
	}

	_, err := s.pool.Exec("orgDelete", orgID)
	return err
}
