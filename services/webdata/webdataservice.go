package webdata

import (
	"context"
	"strconv"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	ErrFilterMissingGroupID = errors.New("address filter missing GroupID")
	ErrAddressMissing       = errors.New("address did not have IPAddress or HostAddress set")
	ErrNoResponses          = errors.New("no responses extracted from webdata")
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

func (s *Service) GetResponses(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, []*am.HTTPResponse, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebDataResponses, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return 0, nil, nil
}

func (s *Service) GetCertificates(ctx context.Context, userContext am.UserContext, filter *am.WebCertificateFilter) (int, []*am.WebCertificate, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebDataCertificates, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return 0, nil, nil
}

func (s *Service) GetSnapshots(ctx context.Context, userContext am.UserContext, filter *am.WebSnapshotFilter) (int, []*am.WebSnapshot, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebDataSnapshots, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return 0, nil, nil
}

// Update webdata in the database, includes serialized dom & snapshot links, all responses and links, and web certificates
// extracted by the web module
func (s *Service) Update(ctx context.Context, userContext am.UserContext, webData *am.WebData) (int, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "create") {
		return 0, am.ErrUserNotAuthorized
	}

	if webData == nil || webData.Address == nil || webData.Address.AddressID == 0 {
		return 0, am.ErrEmptyAddress
	}

	logger := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).
		Int("GroupID", webData.Address.GroupID).
		Int64("AddressID", webData.Address.AddressID).Logger()

	if err := s.addSnapshots(ctx, userContext, logger, webData); err != nil {
		logger.Warn().Err(err).Msg("failed to insert snapshot and serialized dom")
	}

	orgID, err := s.addResponses(ctx, userContext, logger, webData)
	if err != nil {
		return 0, err
	}
	return orgID, nil
}

func (s *Service) addSnapshots(ctx context.Context, userContext am.UserContext, logger zerolog.Logger, webData *am.WebData) error {
	var err error

	oid := webData.Address.OrgID
	gid := webData.Address.GroupID
	aid := webData.Address.AddressID
	_, err = s.pool.ExecEx(ctx, "insertSnapshot", &pgx.QueryExOptions{}, oid, gid, aid, webData.ResponseTimestamp, webData.SerializedDOMLink, webData.SnapshotLink)
	if err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return v
		}
		return err
	}
	return nil
}

func (s *Service) addResponses(ctx context.Context, userContext am.UserContext, logger zerolog.Logger, webData *am.WebData) (int, error) {
	var tx *pgx.Tx
	var err error

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	responseRows, certificateRows := s.buildRows(logger, webData)

	// if responseRows == 0, then we don't have certificates either, so return.
	if len(responseRows) == 0 {
		return 0, ErrNoResponses
	}

	if _, err := tx.Exec(AddResponsesTempTable); err != nil {
		return 0, err
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{AddResponsesTempTableKey}, AddResponsesTempTableColumns, pgx.CopyFromRows(responseRows))
	if err != nil {
		return 0, err
	}

	if copyCount != len(webData.Responses) {
		return 0, am.ErrAddressCopyCount
	}

	if _, err := tx.Exec(AddResponsesTempToStatus); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to add web status")
		}
		return 0, err
	}

	if _, err := tx.Exec(AddResponsesTempToMime); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to add responses to web mime table")
		}
		return 0, err
	}

	if _, err := tx.Exec(AddTempToResponses); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to add web responses")
		}
		return 0, err
	}

	err = tx.Commit()

	if len(certificateRows) > 0 {
		if err := s.addCertificates(ctx, userContext, logger, certificateRows); err != nil {
			logger.Error().Err(err).Msg("failed to add certificates")
		}
	}

	return webData.Address.OrgID, nil
}

func (s *Service) addCertificates(ctx context.Context, userContext am.UserContext, logger zerolog.Logger, certificateRows [][]interface{}) error {
	var tx *pgx.Tx
	var err error

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(AddCertificatesTempTable); err != nil {
		return err
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{AddCertificatesTempTableKey}, AddCertificatesTempTableColumns, pgx.CopyFromRows(certificateRows))
	if err != nil {
		return err
	}

	if copyCount != len(certificateRows) {
		return am.ErrAddressCopyCount
	}

	if _, err := tx.Exec(AddTempToCertificates); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return v
		}
		return err
	}

	return nil
}

func (s *Service) buildRows(logger zerolog.Logger, webData *am.WebData) ([][]interface{}, [][]interface{}) {
	responseRows := make([][]interface{}, 0)
	certificateRows := make([][]interface{}, 0)

	oid := webData.Address.OrgID
	gid := webData.Address.GroupID
	aid := webData.Address.AddressID

	for _, r := range webData.Responses {
		if r == nil {
			continue
		}
		responsePort, err := strconv.Atoi(r.ResponsePort)
		if err != nil {
			responsePort = 0
		}

		requestedPort, err := strconv.Atoi(r.RequestedPort)
		if err != nil {
			requestedPort = 0
		}

		responseRows = append(responseRows, []interface{}{int32(oid), int32(gid), aid, r.ResponseTimestamp, r.IsDocument, r.Scheme, r.IPAddress,
			r.HostAddress, responsePort, requestedPort, r.URL, r.Headers, r.Status, r.StatusText, r.MimeType, r.RawBodyHash, r.RawBodyLink,
		})
		if r.WebCertificate != nil {
			c := r.WebCertificate
			certificateRows = append(certificateRows, []interface{}{int32(oid), int32(gid), r.ResponseTimestamp, r.HostAddress, responsePort, requestedPort,
				c.Protocol, c.KeyExchange, c.KeyExchangeGroup, c.Cipher, c.Mac, c.CertificateId, c.SubjectName, c.SanList, c.Issuer, c.ValidFrom, c.ValidTo, c.CertificateTransparencyCompliance})
		}
	}

	return responseRows, certificateRows
}
