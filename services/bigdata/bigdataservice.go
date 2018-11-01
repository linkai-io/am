package bigdata

import (
	"context"
	"time"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var (
	ErrNoCTRecords = errors.New("no ct records found")
	ErrETLDInvalid = errors.New("etld was empty or did not match")
	ErrCopyCount   = errors.New("count of records copied did not match expected")
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

// GetCT returns locally cached certificate transparency records that match the etld.
func (s *Service) GetCT(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTRecord, error) {
	var emptyTS time.Time
	if !s.IsAuthorized(ctx, userContext, am.RNBigData, "read") {
		return emptyTS, nil, am.ErrUserNotAuthorized
	}

	logger := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).
		Str("ETLD", etld).Logger()

	logger.Info().Msg("processing GetCT request")

	rows, err := s.pool.QueryEx(ctx, "getCertificates", &pgx.QueryExOptions{}, etld)
	if err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return emptyTS, nil, v
		}
		return emptyTS, nil, err
	}
	defer rows.Close()

	records := make(map[string]*am.CTRecord, 0)
	var ts int64
	for rows.Next() {
		r := &am.CTRecord{}

		err := rows.Scan(&ts, &r.CertificateID, &r.InsertedTime, &r.ETLD, &r.CertHash,
			&r.SerialNumber, &r.NotBefore, &r.NotAfter, &r.Country, &r.Organization,
			&r.OrganizationalUnit, &r.CommonName, &r.VerifiedDNSNames, &r.UnverifiedDNSNames,
			&r.IPAddresses, &r.EmailAddresses)

		if err != nil {
			logger.Warn().Err(err).Msg("failed to extract record")
			continue
		}
		records[r.CertHash] = r
	}

	return time.Unix(0, ts), records, err
}

// AddCT adds certificate transparency records
func (s *Service) AddCT(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, ctRecords map[string]*am.CTRecord) error {
	if !s.IsAuthorized(ctx, userContext, am.RNBigData, "create") {
		return am.ErrUserNotAuthorized
	}

	logger := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).
		Str("ETLD", etld).Logger()

	logger.Info().Msg("processing AddCT request")

	var tx *pgx.Tx
	var err error

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	numRecords := len(ctRecords)

	// if numRecords == 0, then we don't have certificates, so return.
	if numRecords == 0 {
		return ErrNoCTRecords
	}

	if etld == "" {
		return ErrETLDInvalid
	}

	if _, err := tx.Exec(AddCTTempTable); err != nil {
		return err
	}

	ctRows := make([][]interface{}, numRecords)
	i := 0
	for _, r := range ctRecords {
		if r == nil || r.ETLD == "" || etld != r.ETLD {
			logger.Warn().Err(ErrETLDInvalid).Str("etld", etld)
			continue
		}

		ctRows[i] = []interface{}{
			r.InsertedTime, r.ETLD, r.CertHash, r.SerialNumber, r.NotBefore, r.NotAfter, r.Country,
			r.Organization, r.OrganizationalUnit, r.CommonName, r.VerifiedDNSNames, r.UnverifiedDNSNames,
			r.IPAddresses, r.EmailAddresses}

		i++
	}

	if _, err := tx.CopyFrom(pgx.Identifier{AddCTTempTableKey}, AddCTTempTableColumns, pgx.CopyFromRows(ctRows)); err != nil {
		return errors.Wrap(err, "copy from for am.certificates failed")
	}

	if _, err := tx.ExecEx(ctx, AddTempToCT, &pgx.QueryExOptions{}); err != nil {
		failedMsg := "failed to add temp certs to am.certificates table"
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, failedMsg)
		}
		return errors.Wrap(err, failedMsg)
	}

	if _, err := tx.ExecEx(ctx, "insertQuery", &pgx.QueryExOptions{}, etld, queryTime.UnixNano()); err != nil {
		failedMsg := "failed to update query time to am.certificate_queries table"
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, failedMsg)
		}
		return errors.Wrap(err, failedMsg)
	}

	return tx.Commit()
}

func (s *Service) DeleteCT(ctx context.Context, userContext am.UserContext, etld string) error {
	if !s.IsAuthorized(ctx, userContext, am.RNBigData, "delete") {
		return am.ErrUserNotAuthorized
	}

	var tx *pgx.Tx
	var err error

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if _, err := tx.ExecEx(ctx, "deleteQuery", &pgx.QueryExOptions{}, etld); err != nil {
		failedMsg := "failed to delete query from am.certificate_queries table"
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, failedMsg)
		}
		return errors.Wrap(err, failedMsg)
	}

	if _, err := tx.ExecEx(ctx, "deleteETLD", &pgx.QueryExOptions{}, etld); err != nil {
		failedMsg := "failed to delete query from am.certificate_queries table"
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, failedMsg)
		}
		return errors.Wrap(err, failedMsg)
	}

	return tx.Commit()
}
