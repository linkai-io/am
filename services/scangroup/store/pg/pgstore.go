package pg

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/am"
)

var (
	ErrEmptyDBAddr       = errors.New("empty db_addr field")
	ErrEmptyDBUser       = errors.New("empty db_user field")
	ErrEmptyDBPass       = errors.New("empty db_pass field")
	ErrEmptyDBName       = errors.New("empty db_name field")
	ErrScanGroupExists   = errors.New("scan group name already exists")
	ErrUserNotAuthorized = errors.New("user is not authorized to perform this action")
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

// Store for interfacing with postgresql/rds
type Store struct {
	pool   *pgx.ConnPool
	config *pgx.ConnPoolConfig
}

// New returns an empty store
func New() *Store {
	return &Store{}
}

// Init by parsing the config and initializing the database pool
func (s *Store) Init(config []byte) error {
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
func (s *Store) parseConfig(config []byte) (*pgx.ConnPoolConfig, error) {
	var v *Config
	if err := json.Unmarshal(config, v); err != nil {
		return nil, err
	}

	if v.Addr == "" {
		return nil, ErrEmptyDBAddr
	}

	if v.Database == "" {
		return nil, ErrEmptyDBName
	}

	if v.User == "" {
		return nil, ErrEmptyDBUser
	}

	if v.Pass == "" {
		return nil, ErrEmptyDBPass
	}

	if v.MaxConnections == 0 {
		v.MaxConnections = 50
	}

	return &pgx.ConnPoolConfig{ConnConfig: pgx.ConnConfig{
		Host:     v.Addr,
		User:     v.User,
		Password: v.Pass,
		Database: v.Database,
	},
		MaxConnections: v.MaxConnections,
		AfterConnect:   s.afterConnect,
	}, nil
}

// afterConnect will iterate over prepared statements with keywords
func (s *Store) afterConnect(conn *pgx.Conn) error {
	for k, v := range queryMap {
		if _, err := conn.Prepare(k, v); err != nil {
			return err
		}
	}
	return nil
}

// IsAuthorized checks if an action is allowed by a particular user
func (s *Store) IsAuthorized(ctx context.Context, orgID, requesterUserID, action am.ScanGroupAction) bool {
	var roleID am.Role
	if err := s.pool.QueryRow("userRole", orgID, requesterUserID).Scan(&roleID); err != nil {
		return false
	}
	if roleID == 0 {
		return false
	}

	return false
}

// Get returns a scan group identified by scangroup id
func (s *Store) Get(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, group *am.ScanGroup, err error) {
	return oid, group, err
}

// Create a new scan group and initial scan group version, returning orgID and groupID on success, error otherwise
func (s *Store) Create(ctx context.Context, newGroup *am.ScanGroup, newVersion *am.ScanGroupVersion) (oid int32, gid int32, err error) {
	var tx *pgx.Tx

	oid = newGroup.OrgID
	// TODO: Do auth check
	err = s.pool.QueryRow("scanGroupIDByName", newGroup.OrgID, newGroup.GroupName).Scan(&gid)
	if err != pgx.ErrNoRows {
		return 0, 0, ErrScanGroupExists
	}

	tx, err = s.pool.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	_, err = tx.Exec("createScanGroup", newGroup.OrgID, newGroup.GroupName, newGroup.CreationTime, newGroup.CreatedBy, newGroup.OriginalInput)
	if err != nil {
		return 0, 0, err
	}

	err = tx.QueryRow("scanGroupIDByName", newVersion.OrgID, newGroup.GroupName).Scan(&gid)
	if err != nil {
		return 0, 0, err
	}

	_, err = tx.Exec("createScanGroupVersion", newVersion.OrgID, gid, newVersion.VersionName, newVersion.CreationTime, newVersion.ModuleConfigurations, newVersion.GroupVersionID)
	if err != nil {
		return 0, 0, err
	}

	err = tx.Commit()
	return oid, gid, err
}

// Delete a scan group, returning orgID and groupID on success, error otherwise
func (s *Store) Delete(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, gid int32, err error) {
	return oid, gid, err
}

// GetVersion returns the configuration of the requested version.
func (s *Store) GetVersion(ctx context.Context, orgID, requesterUserID, groupID, groupVersionID int32) (oid int32, groupVersion *am.ScanGroupVersion, err error) {
	return oid, groupVersion, err
}

// CreateVersion for a scan group, allowing modification of module configurations
func (s *Store) CreateVersion(ctx context.Context, orgID, requesterUserID int32, scanGroupVersion *am.ScanGroupVersion) (oid int32, gid int32, gvid int32, err error) {
	return oid, gid, gvid, err
}

// DeleteVersion requires orgID, groupVersionID and one of groupID or versionName. returning orgID, groupID and groupVersionID if success
func (s *Store) DeleteVersion(ctx context.Context, orgID, requesterUserID, groupID, groupVersionID int32, versionName string) (oid int32, gid int32, gvid int32, err error) {
	return oid, gid, gvid, err
}

// Groups returns all groups for an organization.
func (s *Store) Groups(ctx context.Context, orgID int32) (oid int32, groups []*am.ScanGroup, err error) {
	return oid, groups, err
}

// Addresses returns all addresses for a scan group
func (s *Store) Addresses(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, addresses []*am.ScanGroupAddress, err error) {
	return oid, addresses, err
}
