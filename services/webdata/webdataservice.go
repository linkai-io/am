package webdata

import (
	"context"
	"fmt"
	"strconv"

	"github.com/linkai-io/am/pkg/generators"

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
	ErrCopyCount            = errors.New("count of records copied did not match expected")
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

// GetURLList returns a list of urls for a series of responses (key'd off of urlrequesttimestamp)
func (s *Service) GetURLList(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, []*am.URLListResponse, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebDataResponses, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}

	var getQuery string
	var rows *pgx.Rows
	var args []interface{}
	var err error

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	if filter.Limit > 10000 {
		return 0, nil, am.ErrLimitTooLarge
	}

	if filter.GroupID == 0 {
		return 0, nil, ErrFilterMissingGroupID
	}

	if filter.LatestOnlyValue {
		getQuery, args = s.BuildURLListFilterQuery(userContext, latestOnlyUrlListQueryPrefix, filter)
	} else {
		getQuery, args = s.BuildURLListFilterQuery(userContext, urlListQueryPrefix, filter)
	}

	serviceLog.Info().Str("query", getQuery).Msg("executing query")
	rows, err = s.pool.Query(getQuery, args...)
	defer rows.Close()

	if err != nil {
		return 0, nil, err
	}

	urlLists := make([]*am.URLListResponse, 0)
	for i := 0; rows.Next(); i++ {
		urlList := &am.URLListResponse{}
		var urls [][]byte
		var links []string
		var responseIDs []int64
		var mimeTypes []string
		if err := rows.Scan(&urlList.OrgID, &urlList.GroupID,
			&urlList.URLRequestTimestamp, &urlList.AddressIDHostAddress, &urlList.AddressIDIPAddress,
			&urls, &links, &responseIDs, &mimeTypes); err != nil {

			return 0, nil, err
		}

		if urlList.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}

		// this is terrible TODO Fix
		urlList.URLs = make([]*am.URLData, len(urls))
		for i, url := range urls {
			urlList.URLs[i] = &am.URLData{
				ResponseID:  responseIDs[i],
				URL:         string(url),
				RawBodyLink: links[i],
				MimeType:    mimeTypes[i],
			}
		}
		urlLists = append(urlLists, urlList)
	}

	return userContext.GetOrgID(), urlLists, err
}

func (s *Service) BuildURLListFilterQuery(userContext am.UserContext, query string, filter *am.WebResponseFilter) (string, []interface{}) {
	args := make([]interface{}, 0)

	args = append(args, userContext.GetOrgID())
	args = append(args, filter.GroupID)
	i := 3
	prefix := ""

	if filter.WithResponseTime {
		generators.AppendConditionalQuery(&query, &prefix, "and (wb.url_request_timestamp=0 OR wb.url_request_timestamp > $%d)", filter.SinceResponseTime, &args, &i)
	}
	if filter.LatestOnlyValue {
		query += " group by wb.organization_id, wb.scan_group_id, address_id_host_address, address_id_ip_address, latest.url_request_timestamp"
	} else {
		query += " group by wb.organization_id, wb.scan_group_id, address_id_host_address, address_id_ip_address, wb.url_request_timestamp order by wb.url_request_timestamp"
	}

	return query, args
}

// GetResponses that match the provided filter.
func (s *Service) GetResponses(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, []*am.HTTPResponse, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebDataResponses, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	var getQuery string
	var args []interface{}
	var rows *pgx.Rows
	var err error

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	if filter.Limit > 10000 {
		return 0, nil, am.ErrLimitTooLarge
	}

	if filter.GroupID == 0 {
		return 0, nil, ErrFilterMissingGroupID
	}

	if filter.LatestOnlyValue {
		getQuery, args = s.BuildWebFilterQuery(userContext, latestOnlyResponseQueryPrefix, filter)
	} else {
		getQuery, args = s.BuildWebFilterQuery(userContext, responseQueryPrefix, filter)
	}

	serviceLog.Info().Str("query", getQuery).Msg("executing query")

	rows, err = s.pool.Query(getQuery, args...)
	defer rows.Close()

	if err != nil {
		return 0, nil, err
	}

	responses := make([]*am.HTTPResponse, 0)

	for i := 0; rows.Next(); i++ {
		r := &am.HTTPResponse{}
		var responsePort int
		var requestedPort int
		var url []byte

		if err := rows.Scan(&r.ResponseID, &r.OrgID, &r.GroupID, &r.AddressID, &r.URLRequestTimestamp,
			&r.ResponseTimestamp, &r.IsDocument, &r.Scheme, &r.IPAddress, &r.HostAddress, &responsePort,
			&requestedPort, &url, &r.Headers, &r.Status, &r.StatusText, &r.MimeType, &r.RawBodyHash,
			&r.RawBodyLink, &r.IsDeleted, &r.AddressIDHostAddress, &r.AddressIDIPAddress); err != nil {

			return 0, nil, err
		}

		r.ResponsePort = strconv.Itoa(responsePort)
		r.RequestedPort = strconv.Itoa(requestedPort)
		r.URL = string(url)

		if r.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}

		responses = append(responses, r)
	}

	return userContext.GetOrgID(), responses, err
}

// BuildWebFilterQuery for building a parameterized query that allows for dynamic conditionals. I'll admit, this
// is one hell of a gnarly generator :/.
func (s *Service) BuildWebFilterQuery(userContext am.UserContext, query string, filter *am.WebResponseFilter) (string, []interface{}) {
	args := make([]interface{}, 0)

	args = append(args, userContext.GetOrgID())
	args = append(args, filter.GroupID)
	i := 3
	prefix := ""

	if filter.WithResponseTime {
		generators.AppendConditionalQuery(&query, &prefix, "(response_timestamp=0 OR response_timestamp > $%d)", filter.SinceResponseTime, &args, &i)
	}

	if filter.WithHeader != "" {
		generators.AppendConditionalQuery(&query, &prefix, "headers ? $%d", filter.WithHeader, &args, &i)
	}

	if filter.WithoutHeader != "" {
		generators.AppendConditionalQuery(&query, &prefix, "not(headers ? $%d)", filter.WithoutHeader, &args, &i)
	}

	if filter.MimeType != "" {
		if filter.LatestOnlyValue {
			generators.AppendConditionalQuery(&query, &prefix, "web_responses.mime_type_id=(select mime_type_id from am.web_mime_type where mime_type=$%d)", filter.MimeType, &args, &i)
		} else {
			generators.AppendConditionalQuery(&query, &prefix, "wb.mime_type_id=(select mime_type_id from am.web_mime_type where mime_type=$%d)", filter.MimeType, &args, &i)
		}
	}

	if filter.LatestOnlyValue {
		query += fmt.Sprintf(`group by web_responses.url) as latest 
		inner join am.web_responses as wb on wb.url_request_timestamp=latest.url_request_timestamp
		join am.web_status_text as wst on wb.status_text_id = wst.status_text_id
		join am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id
		where wb.response_id > $%d order by wb.response_id limit $%d`, i, i+1)
		args = append(args, filter.Start)
		args = append(args, filter.Limit)
		return query, args
	}
	query += fmt.Sprintf("%sresponse_id > $%d order by response_id limit $%d", prefix, i, i+1)
	args = append(args, filter.Start)
	args = append(args, filter.Limit)

	return query, args
}

// GetCertificates that match the provided filter.
func (s *Service) GetCertificates(ctx context.Context, userContext am.UserContext, filter *am.WebCertificateFilter) (int, []*am.WebCertificate, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebDataCertificates, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}

	var rows *pgx.Rows
	var err error

	if filter.Limit > 10000 {
		return 0, nil, am.ErrLimitTooLarge
	}

	if filter.GroupID == 0 {
		return 0, nil, ErrFilterMissingGroupID
	}

	if filter.WithResponseTime {
		rows, err = s.pool.Query("certificatesSinceResponseTime", userContext.GetOrgID(), filter.GroupID, filter.SinceResponseTime, filter.Start, filter.Limit)
	} else {
		rows, err = s.pool.Query("certificatesAll", userContext.GetOrgID(), filter.GroupID, filter.Start, filter.Limit)
	}
	defer rows.Close()

	if err != nil {
		return 0, nil, err
	}

	certificates := make([]*am.WebCertificate, 0)

	for i := 0; rows.Next(); i++ {
		w := &am.WebCertificate{}
		var port int
		if err := rows.Scan(&w.CertificateID, &w.OrgID, &w.GroupID, &w.ResponseTimestamp,
			&w.HostAddress, &port, &w.Protocol, &w.KeyExchange, &w.KeyExchangeGroup,
			&w.Cipher, &w.Mac, &w.CertificateValue, &w.SubjectName, &w.SanList, &w.Issuer,
			&w.ValidFrom, &w.ValidTo, &w.CertificateTransparencyCompliance, &w.IsDeleted); err != nil {

			return 0, nil, err
		}

		w.Port = strconv.Itoa(port)

		if w.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}

		certificates = append(certificates, w)
	}

	return userContext.GetOrgID(), certificates, err
}

// GetSnapshots that match the provided filter
func (s *Service) GetSnapshots(ctx context.Context, userContext am.UserContext, filter *am.WebSnapshotFilter) (int, []*am.WebSnapshot, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebDataSnapshots, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	var rows *pgx.Rows
	var err error

	if filter.Limit > 10000 {
		return 0, nil, am.ErrLimitTooLarge
	}

	if filter.GroupID == 0 {
		return 0, nil, ErrFilterMissingGroupID
	}

	if filter.WithResponseTime {
		rows, err = s.pool.Query("snapshotsSinceResponseTime", userContext.GetOrgID(), filter.GroupID, filter.SinceResponseTime, filter.Start, filter.Limit)
	} else {
		rows, err = s.pool.Query("snapshotsAll", userContext.GetOrgID(), filter.GroupID, filter.Start, filter.Limit)
	}
	defer rows.Close()

	if err != nil {
		return 0, nil, err
	}

	snapshots := make([]*am.WebSnapshot, 0)

	for i := 0; rows.Next(); i++ {
		w := &am.WebSnapshot{}
		if err := rows.Scan(&w.SnapshotID, &w.OrgID, &w.GroupID, &w.AddressID, &w.ResponseTimestamp,
			&w.SerializedDOMHash, &w.SerializedDOMLink, &w.SnapshotLink, &w.IsDeleted, &w.AddressIDHostAddress, &w.AddressIDIPAddress); err != nil {

			return 0, nil, err
		}

		if w.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}

		snapshots = append(snapshots, w)
	}

	return userContext.GetOrgID(), snapshots, err
}

// Add webdata to the database, includes serialized dom & snapshot links, all responses and links, and web certificates
// extracted by the web module
func (s *Service) Add(ctx context.Context, userContext am.UserContext, webData *am.WebData) (int, error) {
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
	_, err = s.pool.ExecEx(ctx, "insertSnapshot", &pgx.QueryExOptions{}, oid, gid, aid, webData.ResponseTimestamp, webData.SerializedDOMHash, webData.SerializedDOMLink, webData.SnapshotLink)
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
		return 0, ErrCopyCount
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
		if err := s.addCertificates(ctx, userContext, certificateRows); err != nil {
			logger.Error().Err(err).Msg("failed to add certificates")
		}
	}

	return webData.Address.OrgID, nil
}

func (s *Service) addCertificates(ctx context.Context, userContext am.UserContext, certificateRows [][]interface{}) error {
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
		log.Ctx(ctx).Warn().Err(err).Msg("failed copy from")
		return err
	}

	if copyCount != len(certificateRows) {
		return ErrCopyCount
	}

	if _, err := tx.Exec(AddTempToCertificates); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return v
		}
		return err
	}

	return tx.Commit()
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

		responseRows = append(responseRows, []interface{}{int32(oid), int32(gid), aid, webData.URLRequestTimestamp, r.ResponseTimestamp, r.IsDocument, r.Scheme, r.IPAddress,
			r.HostAddress, responsePort, requestedPort, r.URL, r.Headers, r.Status, r.StatusText, r.MimeType, r.RawBodyHash, r.RawBodyLink,
		})

		if r.WebCertificate != nil {
			c := r.WebCertificate
			certificateRows = append(certificateRows, []interface{}{int32(oid), int32(gid), r.ResponseTimestamp, r.HostAddress, responsePort,
				c.Protocol, c.KeyExchange, c.KeyExchangeGroup, c.Cipher, c.Mac, c.CertificateValue, c.SubjectName, c.SanList, c.Issuer, c.ValidFrom, c.ValidTo, c.CertificateTransparencyCompliance})
		}
	}

	return responseRows, certificateRows
}
