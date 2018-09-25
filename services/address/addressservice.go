package address

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
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

	var rows *pgx.Rows
	if filter.Limit > 10000 {
		return 0, nil, am.ErrLimitTooLarge
	}

	if filter.GroupID == 0 {
		return 0, nil, ErrFilterMissingGroupID
	}

	if filter.WithIgnored {
		rows, err = s.pool.Query("scanGroupAddressesIgnored", userContext.GetOrgID(), filter.GroupID, filter.IgnoredValue, filter.Start, filter.Limit)
	} else if filter.WithLastScannedTime {
		rows, err = s.pool.Query("scanGroupAddressesSinceScannedTime", userContext.GetOrgID(), filter.GroupID, filter.SinceScannedTime, filter.Start, filter.Limit)
	} else {
		rows, err = s.pool.Query("scanGroupAddressesAll", userContext.GetOrgID(), filter.GroupID, filter.Start, filter.Limit)
	}
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

// Update the scan_group_addresses table. Will do upsert for records, allowing updates as well as
// new addresses to be added.
func (s *Service) Update(ctx context.Context, userContext am.UserContext, addresses []*am.ScanGroupAddress) (oid int, count int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNAddressAddresses, "create") {
		return 0, 0, am.ErrUserNotAuthorized
	}
	var tx *pgx.Tx

	log.Printf("adding: %d\n", len(addresses))

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

	for i := 0; i < numAddresses; i++ {
		a := addresses[i]

		if a.HostAddress == "" && a.IPAddress == "" {
			return 0, 0, ErrAddressMissing
		}

		addressRows[i] = []interface{}{a.AddressID, int32(orgID), int32(a.GroupID), a.HostAddress, a.IPAddress,
			a.DiscoveryTime, a.DiscoveredBy, a.LastScannedTime, a.LastSeenTime, a.ConfidenceScore, a.UserConfidenceScore,
			a.IsSOA, a.IsWildcardZone, a.IsHostedService, a.Ignored, a.FoundFrom, a.NSRecord, a.AddressHash,
		}

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
			return 0, 0, fmt.Errorf("%#v", v)
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
