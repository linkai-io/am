package scangroup

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/auth"
)

var (
	ErrEmptyDBConfig          = errors.New("empty database connection string")
	ErrInvalidDBString        = errors.New("invalid db connection string")
	ErrOrgIDMismatch          = errors.New("org id does not user context")
	ErrScanGroupNotExists     = errors.New("scan group name does not exist")
	ErrScanGroupExists        = errors.New("scan group name already exists")
	ErrUserNotAuthorized      = errors.New("user is not authorized to perform this action")
	ErrScanGroupVersionLinked = errors.New("scan group version is linked to this scan group")
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
		return nil, ErrEmptyDBConfig
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		return nil, ErrInvalidDBString
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

// Get returns a scan group identified by scangroup id
func (s *Service) Get(ctx context.Context, userContext am.UserContext, groupID int) (oid int, group *am.ScanGroup, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "read") {
		return 0, nil, ErrUserNotAuthorized
	}
	group = &am.ScanGroup{}

	//organization_id, scan_group_id, scan_group_name, creation_time, created_by, original_input
	err = s.pool.QueryRow("scanGroupByID", userContext.GetOrgID(), groupID).Scan(
		&group.OrgID, &group.GroupID, &group.GroupName, &group.CreationTime, &group.CreatedBy, &group.ModifiedTime, &group.ModifiedBy,
		&group.OriginalInput, &group.ModuleConfigurations, &group.Deleted,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil, ErrScanGroupNotExists
		}
		return 0, nil, err
	}

	if group.OrgID != userContext.GetOrgID() {
		return 0, nil, ErrOrgIDMismatch
	}

	return group.OrgID, group, err
}

// GetByName returns the scan group identified by scangroup name
func (s *Service) GetByName(ctx context.Context, userContext am.UserContext, groupName string) (oid int, group *am.ScanGroup, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "read") {
		return 0, nil, ErrUserNotAuthorized
	}
	group = &am.ScanGroup{}

	err = s.pool.QueryRow("scanGroupByName", userContext.GetOrgID(), groupName).Scan(
		&group.OrgID, &group.GroupID, &group.GroupName, &group.CreationTime, &group.CreatedBy, &group.ModifiedTime, &group.ModifiedBy,
		&group.OriginalInput, &group.ModuleConfigurations, &group.Deleted,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil, ErrScanGroupNotExists
		}
		return 0, nil, err
	}

	if group.OrgID != userContext.GetOrgID() {
		return 0, nil, ErrOrgIDMismatch
	}

	return group.OrgID, group, err
}

// Groups returns all groups for an organization.
func (s *Service) Groups(ctx context.Context, userContext am.UserContext) (oid int, groups []*am.ScanGroup, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "read") {
		return 0, nil, ErrUserNotAuthorized
	}
	rows, err := s.pool.Query("scanGroupsByOrgID", userContext.GetOrgID())
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	groups = make([]*am.ScanGroup, 0)
	for rows.Next() {
		group := &am.ScanGroup{}
		if err := rows.Scan(&group.OrgID, &group.GroupID, &group.GroupName, &group.CreationTime, &group.CreatedBy, &group.ModifiedTime, &group.ModifiedBy, &group.OriginalInput, &group.ModuleConfigurations, &group.Deleted); err != nil {
			return 0, nil, err
		}

		if group.OrgID != userContext.GetOrgID() {
			return 0, nil, ErrOrgIDMismatch
		}

		groups = append(groups, group)
	}
	return userContext.GetOrgID(), groups, err
}

// Create a new scan group, returning orgID and groupID on success, error otherwise
func (s *Service) Create(ctx context.Context, userContext am.UserContext, newGroup *am.ScanGroup) (oid int, gid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "create") {
		return 0, 0, ErrUserNotAuthorized
	}

	err = s.pool.QueryRow("scanGroupIDByName", userContext.GetOrgID(), newGroup.GroupName).Scan(&oid, &gid)
	if err != nil && err != pgx.ErrNoRows {
		return 0, 0, err
	}

	if gid != 0 {
		return 0, 0, ErrScanGroupExists
	}
	log.Printf("%#v\n", newGroup)
	// creates and sets oid/gid
	err = s.pool.QueryRow("createScanGroup", userContext.GetOrgID(), newGroup.GroupName, newGroup.CreationTime, newGroup.CreatedBy, newGroup.ModifiedTime, newGroup.ModifiedBy, newGroup.OriginalInput, newGroup.ModuleConfigurations).Scan(&oid, &gid)
	if err != nil {
		return 0, 0, err
	}

	return oid, gid, err
}

// Update a scan group, returning orgID and groupID on success, error otherwise
func (s *Service) Update(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) (oid int, gid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "update") {
		return 0, 0, ErrUserNotAuthorized
	}

	err = s.pool.QueryRow("updateScanGroup", group.GroupName, group.ModifiedTime, group.ModifiedBy, group.ModuleConfigurations, userContext.GetOrgID(), group.GroupID).Scan(&oid, &gid)
	if err != nil {
		return 0, 0, err
	}

	return oid, gid, err
}

// Delete a scan group, also deletes all scan group versions which reference this scan group returning orgID and groupID on success, error otherwise
func (s *Service) Delete(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "delete") {
		return 0, 0, ErrUserNotAuthorized
	}
	var tx *pgx.Tx
	var name string

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	// get the current group name so we can change it on delete.
	err = tx.QueryRow("scanGroupName", userContext.GetOrgID(), groupID).Scan(&oid, &name)
	if err != nil {
		return 0, 0, err
	}

	// ensure room for timestamp
	if len(name) > 200 {
		name = name[:200]
	}

	name = fmt.Sprintf("%s_%d", name, time.Now().UnixNano())
	_, err = tx.Exec("deleteScanGroup", name, userContext.GetOrgID(), groupID)
	if err != nil {
		return 0, 0, err
	}

	err = tx.Commit()
	return userContext.GetOrgID(), groupID, err
}

// Addresses returns all addresses for a scan group
func (s *Service) Addresses(ctx context.Context, userContext am.UserContext, groupID int) (oid int, addresses []*am.ScanGroupAddress, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupAddresses, "read") {
		return 0, nil, ErrUserNotAuthorized
	}
	return oid, addresses, err
}

func (s *Service) AddAddresses(ctx context.Context, userContext am.UserContext, addresses []*am.ScanGroupAddress) (oid int, failed []*am.FailedAddress, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupAddresses, "create") {
		return 0, nil, ErrUserNotAuthorized
	}
	return oid, failed, err
}

func (s *Service) UpdateAddresses(ctx context.Context, userContext am.UserContext, addresses []*am.ScanGroupAddress) (oid int, failed []*am.FailedAddress, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupAddresses, "update") {
		return 0, nil, ErrUserNotAuthorized
	}
	return oid, failed, err
}
