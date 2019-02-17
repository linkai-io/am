package address

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/generators"
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

	query, args := s.BuildGetFilterQuery(userContext, filter)
	serviceLog.Info().Msgf("Building Get query with filter: %#v", filter)
	rows, err = s.pool.Query(query, args...)
	defer rows.Close()

	if err != nil {
		return 0, nil, err
	}

	addresses = make([]*am.ScanGroupAddress, 0)

	for i := 0; rows.Next(); i++ {
		a := &am.ScanGroupAddress{}
		if err := rows.Scan(&a.OrgID, &a.AddressID, &a.GroupID, &a.HostAddress,
			&a.IPAddress, &a.DiscoveryTime, &a.DiscoveredBy, &a.LastScannedTime,
			&a.LastSeenTime, &a.ConfidenceScore, &a.UserConfidenceScore, &a.IsSOA,
			&a.IsWildcardZone, &a.IsHostedService, &a.Ignored, &a.FoundFrom, &a.NSRecord, &a.AddressHash); err != nil {

			return 0, nil, err
		}

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

	rows, err = s.pool.Query("scanGroupHostList", userContext.GetOrgID(), filter.GroupID, filter.Start, filter.Limit)
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

// BuildGetFilterQuery creates a 'dynamic' (but prepared statement) filter for filtering scan group addresses
func (s *Service) BuildGetFilterQuery(userContext am.UserContext, filter *am.ScanGroupAddressFilter) (string, []interface{}) {
	args := make([]interface{}, 0)

	query := fmt.Sprintf(`select %s from am.scan_group_addresses as sga where organization_id=$1 and scan_group_id=$2 and `, sharedColumns)
	args = append(args, userContext.GetOrgID())
	args = append(args, filter.GroupID)
	i := 3
	prefix := ""
	if filter.WithIgnored {
		generators.AppendConditionalQuery(&query, &prefix, "ignored=$%d", filter.IgnoredValue, &args, &i)
	}

	if filter.WithLastScannedTime {
		generators.AppendConditionalQuery(&query, &prefix, "(last_scanned_timestamp=0 OR last_scanned_timestamp > $%d)", filter.SinceScannedTime, &args, &i)
	}

	if filter.WithLastSeenTime {
		generators.AppendConditionalQuery(&query, &prefix, "(last_seen_timestamp=0 OR last_seen_timestamp > $%d)", filter.SinceSeenTime, &args, &i)
	}

	if filter.WithIsWildcard {
		generators.AppendConditionalQuery(&query, &prefix, "is_wildcard_zone=$%d", filter.IsWildcardValue, &args, &i)
	}

	if filter.WithIsHostedService {
		generators.AppendConditionalQuery(&query, &prefix, "is_hosted_service=$%d", filter.IsHostedServiceValue, &args, &i)
	}

	if filter.MatchesHost != "" {
		generators.AppendConditionalQuery(&query, &prefix, "reverse(host_address) like '%%$%d'", convert.Reverse(filter.MatchesHost), &args, &i)
	}

	if filter.MatchesIP != "" {
		generators.AppendConditionalQuery(&query, &prefix, "reverse(ip_address) like '%%$%d'", convert.Reverse(filter.MatchesIP), &args, &i)
	}

	if filter.NSRecord != 0 {
		generators.AppendConditionalQuery(&query, &prefix, "ns_record=$%d", filter.NSRecord, &args, &i)
	}

	query += fmt.Sprintf("%saddress_id > $%d order by address_id limit $%d", prefix, i, i+1)
	args = append(args, filter.Start)
	args = append(args, filter.Limit)
	return query, args
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
			a.DiscoveryTime, a.DiscoveredBy, a.LastScannedTime, a.LastSeenTime, a.ConfidenceScore, a.UserConfidenceScore,
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

// Delete the address from the scan group (cascading all delete's across tables).
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
