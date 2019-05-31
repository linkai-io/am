package scangroup

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
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
// returns error if no results found for group id
func (s *Service) Get(ctx context.Context, userContext am.UserContext, groupID int) (oid int, group *am.ScanGroup, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	group = &am.ScanGroup{}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("GroupID", groupID).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Retrieving scan group by id")
	var createTime time.Time
	var modifyTime time.Time
	var pausedTime time.Time

	//organization_id, scan_group_id, scan_group_name, creation_time, created_by, original_input
	err = s.pool.QueryRow("scanGroupByID", userContext.GetOrgID(), groupID).Scan(
		&group.OrgID, &group.GroupID, &group.GroupName, &createTime, &group.CreatedBy, &group.CreatedByID, &modifyTime, &group.ModifiedBy, &group.ModifiedByID,
		&group.OriginalInputS3URL, &group.ModuleConfigurations, &group.Paused, &group.Deleted, &pausedTime, &group.ArchiveAfterDays,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil, am.ErrScanGroupNotExists
		}
		return 0, nil, err
	}

	if group.OrgID != userContext.GetOrgID() {
		return 0, nil, am.ErrOrgIDMismatch
	}

	group.CreationTime = createTime.UnixNano()
	group.ModifiedTime = modifyTime.UnixNano()
	group.LastPausedTime = pausedTime.UnixNano()

	return group.OrgID, group, err
}

// GetByName returns the scan group identified by scangroup name
// returns error if no results found for group name
func (s *Service) GetByName(ctx context.Context, userContext am.UserContext, groupName string) (oid int, group *am.ScanGroup, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	group = &am.ScanGroup{}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Str("GroupName", groupName).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Retrieving scan group by name")

	var createTime time.Time
	var modifyTime time.Time
	var pausedTime time.Time

	err = s.pool.QueryRow("scanGroupByName", userContext.GetOrgID(), groupName).Scan(
		&group.OrgID, &group.GroupID, &group.GroupName, &createTime, &group.CreatedBy, &group.CreatedByID, &modifyTime, &group.ModifiedBy, &group.ModifiedByID,
		&group.OriginalInputS3URL, &group.ModuleConfigurations, &group.Paused, &group.Deleted, &pausedTime, &group.ArchiveAfterDays,
	)
	group.CreationTime = createTime.UnixNano()
	group.ModifiedTime = modifyTime.UnixNano()
	group.LastPausedTime = pausedTime.UnixNano()

	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil, am.ErrScanGroupNotExists
		}
		return 0, nil, err
	}

	if group.OrgID != userContext.GetOrgID() {
		return 0, nil, am.ErrOrgIDMismatch
	}

	return group.OrgID, group, err
}

// AllGroups is a system method for returning groups that match the supplied filter.
func (s *Service) AllGroups(ctx context.Context, userContext am.UserContext, groupFilter *am.ScanGroupFilter) (groups []*am.ScanGroup, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupAllGroups, "read") {
		return nil, am.ErrUserNotAuthorized
	}

	var rows *pgx.Rows

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Retrieving All Groups")

	if val, ok := groupFilter.Filters.Bool("paused"); ok {
		serviceLog.Info().Bool("paused", val).Msg("querying with paused")
		rows, err = s.pool.Query("allScanGroupsWithPaused", val)
	} else {
		serviceLog.Info().Msg("querying all scan groups (except deleted")
		rows, err = s.pool.Query("allScanGroups")
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups = make([]*am.ScanGroup, 0)
	for rows.Next() {
		var createTime time.Time
		var modifyTime time.Time
		var pausedTime time.Time

		group := &am.ScanGroup{}
		if err := rows.Scan(&group.OrgID, &group.GroupID, &group.GroupName, &createTime, &group.CreatedBy,
			&group.CreatedByID, &modifyTime, &group.ModifiedBy, &group.ModifiedByID, &group.OriginalInputS3URL,
			&group.ModuleConfigurations, &group.Paused, &group.Deleted, &pausedTime, &group.ArchiveAfterDays); err != nil {
			return nil, err
		}
		group.CreationTime = createTime.UnixNano()
		group.ModifiedTime = modifyTime.UnixNano()
		group.LastPausedTime = pausedTime.UnixNano()

		groups = append(groups, group)
	}

	return groups, nil
}

// Groups returns all groups for an organization.
func (s *Service) Groups(ctx context.Context, userContext am.UserContext) (oid int, groups []*am.ScanGroup, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Retrieving Groups")

	rows, err := s.pool.Query("scanGroupsByOrgID", userContext.GetOrgID())
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	groups = make([]*am.ScanGroup, 0)
	for rows.Next() {
		var createTime time.Time
		var modifyTime time.Time
		var pausedTime time.Time

		group := &am.ScanGroup{}
		if err := rows.Scan(&group.OrgID, &group.GroupID, &group.GroupName, &createTime, &group.CreatedBy,
			&group.CreatedByID, &modifyTime, &group.ModifiedBy, &group.ModifiedByID, &group.OriginalInputS3URL,
			&group.ModuleConfigurations, &group.Paused, &group.Deleted, &pausedTime, &group.ArchiveAfterDays); err != nil {
			return 0, nil, err
		}

		if group.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}

		group.CreationTime = createTime.UnixNano()
		group.ModifiedTime = modifyTime.UnixNano()
		group.LastPausedTime = pausedTime.UnixNano()

		groups = append(groups, group)
	}
	return userContext.GetOrgID(), groups, err
}

// Create a new scan group, returning orgID and groupID on success, error otherwise
func (s *Service) Create(ctx context.Context, userContext am.UserContext, newGroup *am.ScanGroup) (oid int, gid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "create") {
		return 0, 0, am.ErrUserNotAuthorized
	}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Creating Scan group")

	err = s.pool.QueryRow("scanGroupIDByName", userContext.GetOrgID(), newGroup.GroupName).Scan(&oid, &gid)
	if err != nil && err != pgx.ErrNoRows {
		return 0, 0, err
	}

	if gid != 0 {
		return 0, 0, am.ErrScanGroupExists
	}

	if newGroup.ArchiveAfterDays == 0 {
		newGroup.ArchiveAfterDays = am.DefaultArchiveDays
	}
	// creates and sets oid/gid
	err = s.pool.QueryRow("createScanGroup", userContext.GetOrgID(), newGroup.GroupName, time.Unix(0, newGroup.CreationTime),
		newGroup.CreatedByID, time.Unix(0, newGroup.ModifiedTime), newGroup.ModifiedByID, newGroup.OriginalInputS3URL,
		newGroup.ModuleConfigurations, newGroup.ArchiveAfterDays).Scan(&oid, &gid)
	if err != nil {
		return 0, 0, err
	}

	return oid, gid, err
}

// Update a scan group, returning orgID and groupID on success, error otherwise
func (s *Service) Update(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) (oid int, gid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "update") {
		return 0, 0, am.ErrUserNotAuthorized
	}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int("GroupID", group.GroupID).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Updating Scan group")
	if group.ArchiveAfterDays == 0 {
		group.ArchiveAfterDays = am.DefaultArchiveDays
	}
	err = s.pool.QueryRow("updateScanGroup", group.GroupName, time.Unix(0, group.ModifiedTime), group.ModifiedByID, group.ModuleConfigurations, group.ArchiveAfterDays, userContext.GetOrgID(), group.GroupID).Scan(&oid, &gid)
	if err != nil {
		serviceLog.Error().Err(err).Msgf("Updating Scan group failed %s", queryMap["updateScanGroup"])
		return 0, 0, err
	}

	return oid, gid, err
}

// Delete a scan group, returning orgID and groupID on success, error otherwise
func (s *Service) Delete(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "delete") {
		return 0, 0, am.ErrUserNotAuthorized
	}
	var tx *pgx.Tx
	var name string

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int("GroupID", groupID).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Deleting scan group")

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

	name = fmt.Sprintf("%s_%d\n", name, time.Now().UnixNano())

	_, err = tx.Exec("deleteScanGroup", name, userContext.GetOrgID(), groupID)
	if err != nil {
		return 0, 0, err
	}

	err = tx.Commit()
	return userContext.GetOrgID(), groupID, err
}

// Pause the scan group so it does not get executed by the coordinator
func (s *Service) Pause(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "update") {
		return 0, 0, am.ErrUserNotAuthorized
	}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int("GroupID", groupID).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Pausing scan group")

	now := time.Now()
	err = s.pool.QueryRow("pauseScanGroup", now, userContext.GetUserID(), userContext.GetOrgID(), groupID).Scan(&oid, &gid)
	if err != nil {
		return 0, 0, err
	}

	return oid, gid, err
}

// Resume the scan group so it will get executed by the coordinator
func (s *Service) Resume(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "update") {
		return 0, 0, am.ErrUserNotAuthorized
	}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int("GroupID", groupID).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Resuming scan group")

	now := time.Now()
	err = s.pool.QueryRow("resumeScanGroup", now, userContext.GetUserID(), userContext.GetOrgID(), groupID).Scan(&oid, &gid)
	if err != nil {
		return 0, 0, err
	}

	return oid, gid, err
}

// UpdateStats statics of the scan group (how many are active, size of batching being analyzed etc)
func (s *Service) UpdateStats(ctx context.Context, userContext am.UserContext, stats *am.GroupStats) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "update") {
		return 0, am.ErrUserNotAuthorized
	}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int("GroupID", stats.GroupID).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Updating scan group activity")
	_, err = s.pool.ExecEx(ctx, "updateGroupActivity", &pgx.QueryExOptions{}, userContext.GetOrgID(), stats.GroupID, stats.ActiveAddresses, stats.BatchSize, time.Now(), time.Unix(0, stats.BatchStart), time.Unix(0, stats.BatchEnd))
	if err != nil {
		return 0, err
	}

	return userContext.GetOrgID(), err
}

func (s *Service) GroupStats(ctx context.Context, userContext am.UserContext) (oid int, stats map[int]*am.GroupStats, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNScanGroupGroups, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Retrieving All Groups Activity")

	rows, err := s.pool.Query("getOrgGroupActivity", userContext.GetOrgID())
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	stats = make(map[int]*am.GroupStats, 0)
	for rows.Next() {
		var batchStart time.Time
		var batchEnd time.Time
		var lastUpdated time.Time
		stat := &am.GroupStats{}
		if err := rows.Scan(&stat.OrgID, &stat.GroupID, &stat.ActiveAddresses, &stat.BatchSize, &lastUpdated, &batchStart, &batchEnd); err != nil {
			return 0, nil, err
		}

		if stat.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}
		stat.BatchStart = batchStart.UnixNano()
		stat.BatchEnd = batchEnd.UnixNano()
		stat.LastUpdated = lastUpdated.UnixNano()

		stats[stat.GroupID] = stat
	}
	return userContext.GetOrgID(), stats, err
}
