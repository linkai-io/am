package organization

import (
	"context"
	"log"
	"time"

	uuid "github.com/gofrs/uuid"
	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/auth"
)

// Service for interfacing with postgresql/rds
type Service struct {
	pool        *pgx.ConnPool
	config      *pgx.ConnPoolConfig
	authorizer  auth.Authorizer
	roleManager auth.RoleManager
}

// New returns an Organization Service with a role manager and authorizer
func New(roleManager auth.RoleManager, authorizer auth.Authorizer) *Service {
	return &Service{authorizer: authorizer, roleManager: roleManager}
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
func (s *Service) Get(ctx context.Context, userContext am.UserContext, orgName string) (oid int, org *am.Organization, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("orgByName", orgName))
}

// GetByCID organization by organization customer id
func (s *Service) GetByCID(ctx context.Context, userContext am.UserContext, orgCID string) (oid int, org *am.Organization, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationManage, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("orgByCID", orgCID))
}

// GetByID organization by ID, system user only.
func (s *Service) GetByID(ctx context.Context, userContext am.UserContext, orgID int) (oid int, org *am.Organization, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return s.get(ctx, userContext, s.pool.QueryRow("orgByID", orgID))
}

// get executes the scan against the previously created queryrow
func (s *Service) get(ctx context.Context, userContext am.UserContext, row *pgx.Row) (oid int, org *am.Organization, err error) {
	org = &am.Organization{}
	err = row.Scan(&org.OrgID, &org.OrgName, &org.OrgCID, &org.UserPoolID, &org.IdentityPoolID, &org.OwnerEmail, &org.FirstName, &org.LastName, &org.Phone,
		&org.Country, &org.StatePrefecture, &org.Street, &org.Address1, &org.Address2, &org.City, &org.PostalCode, &org.CreationTime, &org.StatusID, &org.Deleted, &org.SubscriptionID)
	return org.OrgID, org, err
}

// List all organizations that match the supplied filter, system users only.
func (s *Service) List(ctx context.Context, userContext am.UserContext, filter *am.OrgFilter) (orgs []*am.Organization, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "read") {
		return nil, am.ErrUserNotAuthorized
	}
	orgs = make([]*am.Organization, 0)

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

		orgs = append(orgs, org)
	}

	return orgs, nil
}

// Create a new organization, and intialize the user + roles, system users only.
func (s *Service) Create(ctx context.Context, userContext am.UserContext, org *am.Organization) (oid int, uid int, ocid string, ucid string, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "create") {
		return 0, 0, "", "", am.ErrUserNotAuthorized
	}
	tx, err := s.pool.Begin()
	defer tx.Rollback()

	name := ""
	err = tx.QueryRow("orgExists", org.OrgName, -1, "").Scan(&oid, &name, &ocid)
	if err != nil && err != pgx.ErrNoRows {
		return 0, 0, "", "", err
	}

	if oid != 0 {
		return 0, 0, "", "", am.ErrOrganizationExists
	}

	id, err := uuid.NewV4()
	if err != nil {
		return 0, 0, "", "", err
	}
	ocid = id.String()

	id, err = uuid.NewV4()
	if err != nil {
		return 0, 0, "", "", err
	}
	ucid = id.String()

	now := time.Now().UnixNano()
	if err = tx.QueryRow("orgCreate", org.OrgName, ocid, org.UserPoolID, org.IdentityPoolID, org.OwnerEmail,
		org.FirstName, org.LastName, org.Phone, org.Country, org.StatePrefecture, org.Street, org.Address1,
		org.Address2, org.City, org.PostalCode, now, org.StatusID, org.SubscriptionID, ucid, org.OwnerEmail,
		org.FirstName, org.LastName, am.UserStatusActive, now).Scan(&oid, &uid); err != nil {

		return 0, 0, "", "", err
	}

	err = tx.Commit()
	if err != nil {
		return 0, 0, "", "", err
	}

	err = s.addRoles(oid, uid)
	if err != nil {
		// must clean up this org since we committed the transaction
		deleteErr := s.forceDelete(ctx, oid)
		if deleteErr != nil {
			log.Printf("unable to delete organization: %s\n", err)
		}
		return 0, 0, "", "", err
	}

	return oid, uid, ocid, ucid, err
}

// addRoles will add each role necessary for the organization.
// We extract the ownerRoleID and add the supplied userID as a
// member to only the owner role.
func (s *Service) addRoles(orgID, userID int) error {
	ownerRoleID := ""
	for _, roleName := range am.DefaultOrgRoles {

		role := &am.Role{
			OrgID:    orgID,
			RoleName: roleName,
		}

		roleID, err := s.roleManager.CreateRole(role)
		if err != nil {
			return err
		}

		if roleName == am.OwnerRole {
			ownerRoleID = roleID
		}
	}

	return s.roleManager.AddMembers(orgID, ownerRoleID, []int{userID})
}

// Update allows the customer to update the details of their organization
func (s *Service) Update(ctx context.Context, userContext am.UserContext, org *am.Organization) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationManage, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	tx, err := s.pool.Begin()
	defer tx.Rollback()

	oid, update, err := s.get(ctx, userContext, tx.QueryRow("orgByID", userContext.GetOrgID()))
	if err != nil {
		return 0, err
	}

	if org.UserPoolID != "" && org.UserPoolID != update.UserPoolID {
		update.UserPoolID = org.UserPoolID
	}

	if org.IdentityPoolID != "" && org.IdentityPoolID != update.IdentityPoolID {
		update.IdentityPoolID = org.IdentityPoolID
	}

	if org.OwnerEmail != "" && org.OwnerEmail != update.OwnerEmail {
		update.OwnerEmail = org.OwnerEmail
	}

	if org.FirstName != "" && org.FirstName != update.FirstName {
		update.FirstName = org.FirstName
	}

	if org.LastName != "" && org.LastName != update.LastName {
		update.LastName = org.LastName
	}

	if org.Phone != "" && org.Phone != update.Phone {
		update.Phone = org.Phone
	}

	if org.Country != "" && org.Country != update.Country {
		update.Country = org.Country
	}

	if org.StatePrefecture != "" && org.StatePrefecture != update.StatePrefecture {
		update.StatePrefecture = org.StatePrefecture
	}

	if org.Street != "" && org.Street != update.Street {
		update.Street = org.Street
	}

	if org.Address1 != "" && org.Address1 != update.Address1 {
		update.Address1 = org.Address1
	}

	if org.Address2 != "" && org.Address2 != update.Address2 {
		update.Address2 = org.Address2
	}

	if org.City != "" && org.City != update.City {
		update.City = org.City
	}

	if org.PostalCode != "" && org.PostalCode != update.PostalCode {
		update.PostalCode = org.PostalCode
	}

	if org.StatusID != 0 && org.StatusID != update.StatusID {
		update.StatusID = org.StatusID
	}

	if org.SubscriptionID != 0 && org.SubscriptionID != update.SubscriptionID {
		update.SubscriptionID = org.SubscriptionID
	}

	_, err = tx.Exec("orgUpdate", update.UserPoolID, update.IdentityPoolID, update.OwnerEmail, update.FirstName,
		update.LastName, update.Phone, update.Country, update.StatePrefecture, update.Street,
		update.Address1, update.Address2, update.City, update.PostalCode, update.StatusID, update.SubscriptionID, oid)
	if err != nil {
		return 0, err
	}
	return oid, tx.Commit()
}

// Delete the organization
func (s *Service) Delete(ctx context.Context, userContext am.UserContext, orgID int) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNOrganizationSystem, "delete") {
		return 0, am.ErrUserNotAuthorized
	}

	err = s.pool.QueryRow("orgDelete", orgID).Scan(&oid)
	return oid, err
}

// forceDelete must only be called when a critical error occurs while
// creating an organization. While most failures will be caught
// in the transaction rollback, adding roles can not, so we must
// be able to remove the faulty organization.
func (s *Service) forceDelete(ctx context.Context, orgID int) error {
	_, err := s.pool.Exec("orgForceDelete", orgID)
	return err
}
