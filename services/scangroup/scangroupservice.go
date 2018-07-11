package scangroup

import (
	"context"
	"errors"
	"log"

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

// Config represents this modules configuration data to be passed in on
// initialization.
type Config struct {
	Addr           string `json:"db_addr"`
	User           string `json:"db_user"`
	Pass           string `json:"db_pass"`
	Database       string `json:"db_name"`
	MaxConnections int    `json:"db_max_conn"`
}

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
func (s *Service) Get(ctx context.Context, userContext am.UserContext, groupID int32) (oid int32, group *am.ScanGroup, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "read") {
		return 0, nil, ErrUserNotAuthorized
	}

	//organization_id, scan_group_id, scan_group_name, creation_time, created_by, original_input
	err = s.pool.QueryRow("scanGroupIDByName", userContext.GetOrgID(), groupID).Scan(&group.OrgID, &group.GroupID, &group.GroupName, &group.CreationTime, &group.CreatedBy, &group.Deleted)
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
func (s *Service) Groups(ctx context.Context, userContext am.UserContext) (oid int32, groups []*am.ScanGroup, err error) {
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
		if err := rows.Scan(&group.OrgID, &group.GroupID, &group.GroupName, &group.CreationTime, &group.CreatedBy, &group.Deleted); err != nil {
			return 0, nil, err
		}

		if group.OrgID != userContext.GetOrgID() {
			return 0, nil, ErrOrgIDMismatch
		}

		groups = append(groups, group)
	}
	return userContext.GetOrgID(), groups, err
}

// Create a new scan group and initial scan group version, returning orgID and groupID on success, error otherwise
func (s *Service) Create(ctx context.Context, userContext am.UserContext, newGroup *am.ScanGroup, newVersion *am.ScanGroupVersion) (oid int32, gid int32, gvid int32, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "create") {
		return 0, 0, 0, ErrUserNotAuthorized
	}

	var tx *pgx.Tx
	oid = newGroup.OrgID

	err = s.pool.QueryRow("scanGroupIDByName", userContext.GetOrgID(), newGroup.GroupName).Scan(&oid, &gid)
	if err != nil && err != pgx.ErrNoRows {
		return 0, 0, 0, err
	}

	if gid != 0 {
		return 0, 0, 0, ErrScanGroupExists
	}

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, 0, 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	err = tx.QueryRow("createScanGroup", userContext.GetOrgID(), newGroup.GroupName, newGroup.CreationTime, newGroup.CreatedBy, newGroup.OriginalInput).Scan(&gid)
	if err != nil {
		return 0, 0, 0, err
	}

	err = tx.QueryRow("createScanGroupVersion", userContext.GetOrgID(), gid, newVersion.VersionName, newVersion.CreationTime, userContext.GetUserID(), newVersion.ModuleConfigurations).Scan(&gvid)
	if err != nil {
		return 0, 0, 0, err
	}

	err = tx.Commit()
	return oid, gid, gvid, err
}

// Delete a scan group, also deletes all scan group versions which reference this scan group returning orgID and groupID on success, error otherwise
func (s *Service) Delete(ctx context.Context, userContext am.UserContext, groupID int32) (oid int32, gid int32, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "delete") {
		return 0, 0, ErrUserNotAuthorized
	}
	var tx *pgx.Tx
	var row *pgx.Row
	versionID := 0

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	// ensure that we have no rows for this scan group version linked to the group id we are about to delete
	row = tx.QueryRow("scanGroupVersionExists", userContext.GetOrgID(), groupID)

	err = row.Scan(&oid, &versionID)
	log.Printf("%d %d\n", oid, versionID)
	if versionID != 0 {
		return 0, 0, ErrScanGroupVersionLinked
	}

	_, err = tx.Exec("deleteScanGroup", userContext.GetOrgID(), groupID)
	if err != nil {
		return 0, 0, err
	}

	err = tx.Commit()
	return userContext.GetOrgID(), groupID, err
}

// GetVersion returns the configuration of the requested version.
func (s *Service) GetVersion(ctx context.Context, userContext am.UserContext, groupID, groupVersionID int32) (oid int32, v *am.ScanGroupVersion, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupVersions, "read") {
		return 0, nil, ErrUserNotAuthorized
	}
	var row *pgx.Row

	row = s.pool.QueryRow("scanGroupVersionByID", userContext.GetOrgID(), groupID, groupVersionID)
	err = row.Scan(&v.OrgID, &v.GroupID, &v.GroupVersionID, &v.VersionName, &v.CreationTime, &v.CreatedBy, &v.ModuleConfigurations, &v.Deleted)

	return v.OrgID, v, err
}

// GetVersionByName returns the configuration of the requested version.
func (s *Service) GetVersionByName(ctx context.Context, userContext am.UserContext, groupID int32, versionName string) (oid int32, v *am.ScanGroupVersion, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupVersions, "read") {
		return 0, nil, ErrUserNotAuthorized
	}
	var row *pgx.Row
	v = &am.ScanGroupVersion{}

	//organization_id, scan_group_id, version_name, creation_time, created_by, configuration, config_version, deleted
	row = s.pool.QueryRow("scanGroupVersionByName", userContext.GetOrgID(), groupID, versionName)
	err = row.Scan(&v.OrgID, &v.GroupID, &v.GroupVersionID, &v.VersionName, &v.CreationTime, &v.CreatedBy, &v.ModuleConfigurations, &v.Deleted)

	return v.OrgID, v, err
}

// CreateVersion for a scan group, allowing modification of module configurations
func (s *Service) CreateVersion(ctx context.Context, userContext am.UserContext, v *am.ScanGroupVersion) (oid int32, gid int32, gvid int32, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupVersions, "create") {
		return 0, 0, 0, ErrUserNotAuthorized
	}
	var row *pgx.Row
	var returned am.ScanGroupVersion
	//organization_id, scan_group_id, version_name, creation_time, created_by, configuration, deleted
	_, err = s.pool.Exec("createScanGroupVersion", v.OrgID, v.GroupID, v.VersionName, v.CreationTime, v.CreatedBy, v.ModuleConfigurations)
	if err != nil {
		return 0, 0, 0, err
	}

	row = s.pool.QueryRow("scanGroupVersionIDs", v.OrgID, v.GroupID)
	err = row.Scan(&returned.OrgID, &returned.GroupID, &returned.GroupVersionID)
	if err != nil {
		return 0, 0, 0, err
	}
	return returned.OrgID, returned.GroupID, returned.GroupVersionID, err
}

// DeleteVersion requires orgID, groupVersionID and one of groupID or versionName. returning orgID, groupID and groupVersionID if success
func (s *Service) DeleteVersion(ctx context.Context, userContext am.UserContext, groupID, groupVersionID int32, versionName string) (oid int32, gid int32, gvid int32, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupVersions, "delete") {
		return 0, 0, 0, ErrUserNotAuthorized
	}
	log.Printf("deleteScanGroupVersion %d %d %d\n", userContext.GetOrgID(), groupID, groupVersionID)
	if t, err := s.pool.Exec("deleteScanGroupVersion", userContext.GetOrgID(), groupID, groupVersionID); err != nil {
		return 0, 0, 0, err
	} else {
		log.Printf("%s\n", t)
	}

	return userContext.GetOrgID(), groupID, groupVersionID, err
}

// Addresses returns all addresses for a scan group
func (s *Service) Addresses(ctx context.Context, userContext am.UserContext, groupID int32) (oid int32, addresses []*am.ScanGroupAddress, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupAddresses, "read") {
		return 0, nil, ErrUserNotAuthorized
	}
	return oid, addresses, err
}

func (s *Service) AddAddresses(ctx context.Context, userContext am.UserContext, addresses []*am.ScanGroupAddress) (oid int32, failed []*am.FailedAddress, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupAddresses, "create") {
		return 0, nil, ErrUserNotAuthorized
	}
	return oid, failed, err
}

func (s *Service) UpdateAddresses(ctx context.Context, userContext am.UserContext, addresses []*am.ScanGroupAddress) (oid int32, failed []*am.FailedAddress, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupAddresses, "update") {
		return 0, nil, ErrUserNotAuthorized
	}
	return oid, failed, err
}
