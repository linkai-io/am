package address

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
)

const (
	sevenDays = time.Hour * time.Duration(-168)
)

var (
	ErrFilterMissingGroupID = errors.New("address filter missing GroupID")
	ErrAddressMissing       = errors.New("address did not have IPAddress or HostAddress set")
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
			log.Error().Err(err).Str("key", k).Msg("failed to init key")
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

// Get returns all addresses for a scan group that match the supplied filter
func (s *Service) Get(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, addresses []*am.ScanGroupAddress, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	serviceLog.Info().Msg("getting address list")

	var rows *pgx.Rows
	if filter.Limit > 10000 {
		return 0, nil, am.ErrLimitTooLarge
	}

	if filter.GroupID == 0 {
		return 0, nil, ErrFilterMissingGroupID
	}

	query, args, err := buildGetFilterQuery(userContext, filter)
	if err != nil {
		return 0, nil, err
	}

	serviceLog.Info().Msgf("Building Get query with filter: %#v %#v", filter, filter.Filters)
	serviceLog.Info().Msgf("%s", query)
	serviceLog.Info().Msgf("%#v", args)
	rows, err = s.pool.Query(query, args...)
	defer rows.Close()

	if err != nil {
		return 0, nil, err
	}

	addresses = make([]*am.ScanGroupAddress, 0)

	for i := 0; rows.Next(); i++ {
		var dTime time.Time
		var scanTime time.Time
		var seenTime time.Time

		a := &am.ScanGroupAddress{}

		if err := rows.Scan(&a.OrgID, &a.AddressID, &a.GroupID, &a.HostAddress,
			&a.IPAddress, &dTime, &a.DiscoveredBy, &scanTime,
			&seenTime, &a.ConfidenceScore, &a.UserConfidenceScore, &a.IsSOA,
			&a.IsWildcardZone, &a.IsHostedService, &a.Ignored, &a.FoundFrom, &a.NSRecord, &a.AddressHash); err != nil {

			return 0, nil, err
		}

		a.DiscoveryTime = dTime.UnixNano()
		a.LastScannedTime = scanTime.UnixNano()
		a.LastSeenTime = seenTime.UnixNano()

		if a.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}

		addresses = append(addresses, a)
	}

	return userContext.GetOrgID(), addresses, err
}

// GetHostList returns hostnames and a list of IP addresses for each host
// TODO: add filtering for start/limit
func (s *Service) GetHostList(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, hosts []*am.ScanGroupHostList, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	serviceLog.Info().Msg("getting host list")

	var rows *pgx.Rows
	if filter.Limit > 10000 {
		return 0, nil, am.ErrLimitTooLarge
	}

	if filter.GroupID == 0 {
		return 0, nil, ErrFilterMissingGroupID
	}

	var startHost string
	var ok bool
	if startHost, ok = filter.Filters.String(am.FilterStartsHostAddress); !ok {
		startHost = ""
	}

	rows, err = s.pool.Query("scanGroupHostList", userContext.GetOrgID(), filter.GroupID, startHost, filter.Limit)
	defer rows.Close()

	if err != nil {
		return 0, nil, err
	}

	hosts = make([]*am.ScanGroupHostList, 0)

	for i := 0; rows.Next(); i++ {
		h := &am.ScanGroupHostList{}
		if err := rows.Scan(&h.OrgID, &h.GroupID, &h.HostAddress, &h.IPAddresses, &h.AddressIDs); err != nil {
			return 0, nil, err
		}

		if h.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}

		hosts = append(hosts, h)
	}

	return userContext.GetOrgID(), hosts, err
}

// Update or insert new addresses
func (s *Service) Update(ctx context.Context, userContext am.UserContext, addresses map[string]*am.ScanGroupAddress) (oid int, count int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "create") {
		return 0, 0, am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	ctx = serviceLog.WithContext(ctx)

	var tx *pgx.Tx

	log.Ctx(ctx).Info().Int("address_len", len(addresses)).Msg("adding")

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if _, err := tx.Exec(AddAddressesTempTable); err != nil {
		return 0, 0, err
	}

	numAddresses := len(addresses)

	addressRows := make([][]interface{}, numAddresses)
	orgID := userContext.GetOrgID()
	i := 0

	for _, a := range addresses {
		if a == nil || (a.HostAddress == "" && a.IPAddress == "") {
			return 0, 0, ErrAddressMissing
		}

		addressRows[i] = []interface{}{int32(orgID), int32(a.GroupID), a.HostAddress, a.IPAddress,
			time.Unix(0, a.DiscoveryTime), a.DiscoveredBy, time.Unix(0, a.LastScannedTime), time.Unix(0, a.LastSeenTime), a.ConfidenceScore, a.UserConfidenceScore,
			a.IsSOA, a.IsWildcardZone, a.IsHostedService, a.Ignored, a.FoundFrom, a.NSRecord, a.AddressHash,
		}

		i++
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{AddAddressesTempTableKey}, AddAddressesTempTableColumns, pgx.CopyFromRows(addressRows))
	if err != nil {
		return 0, 0, err
	}

	if copyCount != numAddresses {
		return 0, 0, am.ErrAddressCopyCount
	}

	if _, err := tx.Exec(AddAddressesTempToAddress); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, 0, errors.Wrap(v, "failed to update addresses")
		}
		return 0, 0, err
	}

	err = tx.Commit()

	return orgID, copyCount, err
}

// Delete the address from the scan group by setting the deleted column to true
func (s *Service) Delete(ctx context.Context, userContext am.UserContext, groupID int, addressIDs []int64) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "delete") {
		return 0, am.ErrUserNotAuthorized
	}
	var tx *pgx.Tx

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	log.Ctx(ctx).Info().Int("address_len", len(addressIDs)).Msg("deleting")

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if _, err := tx.Exec(DeleteAddressesTempTable); err != nil {
		return 0, err
	}

	numAddresses := len(addressIDs)

	addressRows := make([][]interface{}, numAddresses)
	orgID := userContext.GetOrgID()

	for i := 0; i < len(addressIDs); i++ {
		addressRows[i] = []interface{}{addressIDs[i]}
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{DeleteAddressesTempTableKey}, DeleteAddressesTempTableColumns, pgx.CopyFromRows(addressRows))
	if err != nil {
		return 0, err
	}

	if copyCount != numAddresses {
		return 0, am.ErrAddressCopyCount
	}

	if _, err := tx.Exec(DeleteAddressesTempToAddress, orgID, groupID); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to delete addresses")
		}
		return 0, err
	}

	// TODO: instead of a trigger, just automatically unset max hosts for this org. While not great, the next update/insert
	// will test it anyways
	if _, err := tx.Exec("unsetMaxHosts", orgID); err != nil {
		return 0, err
	}

	err = tx.Commit()

	return orgID, err
}

// Ignore the addresses from the scan group.
func (s *Service) Ignore(ctx context.Context, userContext am.UserContext, groupID int, addressIDs []int64, value bool) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "create") {
		return 0, am.ErrUserNotAuthorized
	}
	var tx *pgx.Tx

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	log.Ctx(ctx).Info().Int("address_len", len(addressIDs)).Msg("ignoring")

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if _, err := tx.Exec(IgnoreAddressesTempTable); err != nil {
		return 0, err
	}

	numAddresses := len(addressIDs)

	addressRows := make([][]interface{}, numAddresses)
	orgID := userContext.GetOrgID()

	for i := 0; i < len(addressIDs); i++ {
		addressRows[i] = []interface{}{addressIDs[i]}
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{IgnoreAddressesTempTableKey}, IgnoreAddressesTempTableColumns, pgx.CopyFromRows(addressRows))
	if err != nil {
		return 0, err
	}

	if copyCount != numAddresses {
		return 0, am.ErrAddressCopyCount
	}

	if _, err := tx.Exec(IgnoreAddressesTempToAddress, value, orgID, groupID); err != nil {
		return 0, err
	}

	// TODO: instead of a trigger, just automatically unset max hosts for this org. While not great, the next update/insert
	// will test it anyways
	if _, err := tx.Exec("unsetMaxHosts", orgID); err != nil {
		return 0, err
	}

	err = tx.Commit()

	return orgID, err
}

// Count returns the number of addresses for a specified scan group by id
func (s *Service) Count(ctx context.Context, userContext am.UserContext, groupID int) (oid int, count int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "read") {
		return 0, 0, am.ErrUserNotAuthorized
	}

	err = s.pool.QueryRow("scanGroupAddressesCount", userContext.GetOrgID(), groupID).Scan(&count)
	if err != nil {
		return 0, 0, err
	}

	return userContext.GetOrgID(), count, err
}

func (s *Service) OrgStats(ctx context.Context, userContext am.UserContext) (oid int, orgStats []*am.ScanGroupAddressStats, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	orgStats = make([]*am.ScanGroupAddressStats, 0)

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	rows, err := s.pool.QueryEx(ctx, "discoveredOrgAgg", &pgx.QueryExOptions{}, userContext.GetOrgID(), time.Now().Add(sevenDays))
	defer rows.Close()
	if err != nil {
		return 0, nil, err
	}
	stats := make(map[int]*am.ScanGroupAddressStats, 0)

	for rows.Next() {
		var ok bool
		stat := &am.ScanGroupAddressStats{}
		agg := &am.ScanGroupAggregates{}
		var aggType string
		var groupID int
		var periodStart time.Time
		var count int32
		if err := rows.Scan(&aggType, &groupID, &periodStart, &count); err != nil {
			return 0, nil, err
		}

		if stat, ok = stats[groupID]; !ok {
			stat = &am.ScanGroupAddressStats{}
			stat.DiscoveredBy = make([]string, 0)
			stat.DiscoveredByCount = make([]int32, 0)
			stat.Aggregates = make(map[string]*am.ScanGroupAggregates)
			stat.GroupID = groupID
			stat.OrgID = userContext.GetOrgID()
			stats[groupID] = stat
		}

		if agg, ok = stat.Aggregates[aggType]; !ok {
			agg = &am.ScanGroupAggregates{}
			stat.Aggregates[aggType] = agg
		}

		stat.Aggregates[aggType].Count = append(stat.Aggregates[aggType].Count, count)
		stat.Aggregates[aggType].Time = append(stat.Aggregates[aggType].Time, periodStart.UnixNano())
	}

	rows, err = s.pool.QueryEx(ctx, "discoveredByOrg", &pgx.QueryExOptions{}, userContext.GetOrgID())
	defer rows.Close()
	if err != nil {
		return 0, nil, err
	}

	for rows.Next() {
		stat := &am.ScanGroupAddressStats{}
		var groupID int
		var by string
		var count int32
		var ok bool

		if err := rows.Scan(&groupID, &by, &count); err != nil {
			return 0, nil, err
		}

		if stat, ok = stats[groupID]; !ok {
			stat = &am.ScanGroupAddressStats{}
			stat.Aggregates = make(map[string]*am.ScanGroupAggregates)
			stat.GroupID = groupID
			stat.OrgID = userContext.GetOrgID()
			stats[groupID] = stat
		}

		if stats[groupID].DiscoveredBy == nil {
			stats[groupID].DiscoveredBy = make([]string, 0)
			stats[groupID].DiscoveredByCount = make([]int32, 0)
		}

		stats[groupID].DiscoveredBy = append(stats[groupID].DiscoveredBy, by)
		stats[groupID].DiscoveredByCount = append(stats[groupID].DiscoveredByCount, count)
		stats[groupID].ConfidentTotal += count // since our query already counts only confident hosts, just add to the total
	}

	for _, stat := range stats {
		orgStats = append(orgStats, stat)
	}
	return userContext.GetOrgID(), orgStats, err
}

// GroupStats is lazy and just does the whole org and returns the groupID if it exists... TODO: do it properly.
func (s *Service) GroupStats(ctx context.Context, userContext am.UserContext, groupID int) (oid int, groupStats *am.ScanGroupAddressStats, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}

	oid, orgStats, err := s.OrgStats(ctx, userContext)
	if err != nil {
		return 0, nil, err
	}

	for _, groupStats = range orgStats {
		if groupStats.GroupID == groupID {
			return oid, groupStats, nil
		}
	}
	return userContext.GetOrgID(), nil, am.ErrScanGroupNotExists
}

// Archive records for a group.
func (s *Service) Archive(ctx context.Context, userContext am.UserContext, group *am.ScanGroup, archiveTime time.Time) (oid int, count int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "update") {
		return 0, 0, am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("call", "AddressService.Archive").
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	// double check group pause time
	days := time.Hour * time.Duration(24*group.ArchiveAfterDays)
	archiveBefore := archiveTime.Add(-days)
	pauseTime := time.Unix(0, group.LastPausedTime)
	// if the group has been paused we should archive records that match before pausedTime - days
	if pauseTime.After(archiveBefore) {
		archiveBefore = pauseTime.Add(-days)
	}
	serviceLog.Info().Time("archive_before", archiveBefore).Msg("Archiving records that match filter before")
	// run query against addresses
	// we don't want to archive input_list addresses (maybe NS/MX???)
	tx, err := s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	log.Info().Msgf("QUERY: %s\n", queryMap["archiveHosts"])
	log.Info().Msgf("ARGS: %v %v %v\n", userContext.GetOrgID(), group.GroupID, archiveBefore)
	_, err = tx.ExecEx(ctx, "archiveHosts", &pgx.QueryExOptions{}, userContext.GetOrgID(), group.GroupID, archiveBefore)
	if err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, 0, errors.Wrap(v, "failed to archive addresses")
		}
		return 0, 0, err
	}

	err = tx.Commit()
	// report how many were archived
	return userContext.GetOrgID(), 0, err
}

// UpdateHostPorts saves new port scan results
func (s *Service) UpdateHostPorts(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, portResults *am.PortResults) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "update") {
		return 0, am.ErrUserNotAuthorized
	}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("call", "AddressService.UpdateHostPorts").
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	tx, err := s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if address != nil {
		a := address
		// TODO: FIX HERE ARGUMENT ORDER INCORRECT.
		if _, err := tx.ExecEx(ctx, "insertPortHost", &pgx.QueryExOptions{}, int32(a.OrgID), int32(a.GroupID), a.HostAddress, a.IPAddress,
			time.Unix(0, a.DiscoveryTime), a.DiscoveredBy, time.Unix(0, a.LastScannedTime), time.Unix(0, a.LastSeenTime), a.ConfidenceScore, a.UserConfidenceScore,
			a.IsSOA, a.IsWildcardZone, a.IsHostedService, a.Ignored, a.FoundFrom, a.NSRecord, a.AddressHash); err != nil {
			if v, ok := err.(pgx.PgError); ok {
				return 0, errors.Wrap(v, "failed to insert host from portscan")
			}
			return 0, err
		}
	}

	portResults.Ports.Previous = &am.PortData{} // nil it out so we don't add garbage to the table
	if _, err := tx.ExecEx(ctx, "updateHostPorts", &pgx.QueryExOptions{}, userContext.GetOrgID(), portResults.GroupID, portResults.HostAddress, portResults.Ports, time.Unix(0, portResults.ScannedTimestamp)); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to insert ports")
		}
		return 0, err
	}
	err = tx.Commit()
	return oid, err
}
