package webdata

import (
	"context"
	"strconv"
	"time"

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

func (s *Service) OrgStats(ctx context.Context, userContext am.UserContext) (oid int, orgStats []*am.ScanGroupWebDataStats, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebDataResponses, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Str("call", "webdataservice.OrgStats").
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	log.Ctx(ctx).Info().Msg("getting server counts")
	rows, err := s.pool.QueryEx(ctx, "serverCounts", &pgx.QueryExOptions{}, userContext.GetOrgID())
	if err != nil {
		return 0, nil, err
	}
	stats := make(map[int]*am.ScanGroupWebDataStats)

	defer rows.Close()
	for rows.Next() {
		stat := &am.ScanGroupWebDataStats{}
		var server *string
		var count int32
		var groupID int
		var ok bool

		if err := rows.Scan(&groupID, &server, &count); err != nil {
			return 0, nil, err
		}

		if stat, ok = stats[groupID]; !ok {
			stat = &am.ScanGroupWebDataStats{}
			stat.ServerTypes = make([]string, 0)
			stat.ServerCounts = make([]int32, 0)
			stats[groupID] = stat
		}
		stat.GroupID = groupID
		stat.OrgID = userContext.GetOrgID()
		if server == nil {
			continue
		}
		stat.ServerTypes = append(stat.ServerTypes, *server)
		stat.ServerCounts = append(stat.ServerCounts, count)
		stat.UniqueWebServers += count // add the total of server types (since they are unique host/port)
	}
	log.Ctx(ctx).Info().Msg("got group stats")

	rows, err = s.pool.QueryEx(ctx, "expiringCerts", &pgx.QueryExOptions{}, userContext.GetOrgID())
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		stat := &am.ScanGroupWebDataStats{}

		var ok bool
		var expiresTime string
		var groupID int
		var count int32

		if err := rows.Scan(&groupID, &expiresTime, &count); err != nil {
			return 0, nil, err
		}

		if stat, ok = stats[groupID]; !ok {
			stat = &am.ScanGroupWebDataStats{}
			stat.ServerTypes = make([]string, 0)
			stat.ServerCounts = make([]int32, 0)
			stats[groupID] = stat
		}
		switch expiresTime {
		case "thirty":
			stats[groupID].ExpiringCerts30Days += count
		case "fifteen":
			stats[groupID].ExpiringCerts15Days += count
		}
	}

	for _, v := range stats {
		orgStats = append(orgStats, v)
	}
	log.Ctx(ctx).Info().Msg("got expiring certs stats, returning org stats")
	return userContext.GetOrgID(), orgStats, nil
}

func (s *Service) GroupStats(ctx context.Context, userContext am.UserContext, groupID int) (oid int, groupStats *am.ScanGroupWebDataStats, err error) {
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

	getQuery, args, err = buildURLListFilterQuery(userContext, filter)
	if err != nil {
		return 0, nil, err
	}

	serviceLog.Info().Str("query", getQuery).Msg("executing query")
	log.Info().Msgf("ARGS: %#v\n", args)
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
		var urlRequestTime time.Time

		if err := rows.Scan(&urlList.OrgID, &urlList.GroupID,
			&urlRequestTime, &urlList.HostAddress, &urlList.IPAddress,
			&urls, &links, &responseIDs, &mimeTypes); err != nil {

			return 0, nil, err
		}

		if urlList.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}
		urlList.URLRequestTimestamp = urlRequestTime.UnixNano()

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

	getQuery, args, err = buildWebFilterQuery(userContext, filter)
	if err != nil {
		return 0, nil, err
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
		var responseTime time.Time
		var urlRequestTime time.Time
		var responsePort int
		var requestedPort int
		var url []byte

		if err := rows.Scan(&r.ResponseID, &r.OrgID, &r.GroupID, &r.AddressHash, &urlRequestTime,
			&responseTime, &r.IsDocument, &r.Scheme, &r.IPAddress, &r.HostAddress, &r.LoadIPAddress, &r.LoadHostAddress, &responsePort,
			&requestedPort, &url, &r.Headers, &r.Status, &r.StatusText, &r.MimeType, &r.RawBodyHash,
			&r.RawBodyLink, &r.IsDeleted); err != nil {

			return 0, nil, err
		}
		r.ResponseTimestamp = responseTime.UnixNano()
		r.URLRequestTimestamp = urlRequestTime.UnixNano()
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
	query, args, err := buildCertificateFilter(userContext, filter)
	if err != nil {
		return 0, nil, err
	}

	rows, err = s.pool.Query(query, args...)
	defer rows.Close()

	if err != nil {
		return 0, nil, err
	}

	certificates := make([]*am.WebCertificate, 0)

	for i := 0; rows.Next(); i++ {
		w := &am.WebCertificate{}
		var port int
		var responseTime time.Time
		if err := rows.Scan(&w.CertificateID, &w.OrgID, &w.GroupID, &responseTime, &w.AddressHash,
			&w.HostAddress, &w.IPAddress, &port, &w.Protocol, &w.KeyExchange, &w.KeyExchangeGroup,
			&w.Cipher, &w.Mac, &w.CertificateValue, &w.SubjectName, &w.SanList, &w.Issuer,
			&w.ValidFrom, &w.ValidTo, &w.CertificateTransparencyCompliance, &w.IsDeleted); err != nil {

			return 0, nil, err
		}

		w.Port = strconv.Itoa(port)
		w.ResponseTimestamp = responseTime.UnixNano()
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

	query, args, err := buildSnapshotQuery(userContext, filter)
	if err != nil {
		return 0, nil, err
	}
	log.Info().Msgf("%s %#v", query, args)
	rows, err = s.pool.Query(query, args...)
	defer rows.Close()
	if err != nil {
		return 0, nil, err
	}

	snapshots := make([]*am.WebSnapshot, 0)

	for i := 0; rows.Next(); i++ {
		w := &am.WebSnapshot{}
		var responseTime time.Time
		var urlRequestTime time.Time
		var url []byte
		var loadURL []byte

		if err := rows.Scan(&w.SnapshotID, &w.OrgID, &w.GroupID, &w.AddressHash, &w.HostAddress, &w.IPAddress, &w.Scheme, &w.ResponsePort, &w.RequestedPort, &url, &responseTime,
			&w.SerializedDOMHash, &w.SerializedDOMLink, &w.SnapshotLink, &w.IsDeleted, &loadURL, &urlRequestTime,
			&w.TechCategories, &w.TechNames, &w.TechVersions, &w.TechMatchLocations, &w.TechMatchData, &w.TechIcons, &w.TechWebsites); err != nil {
			return 0, nil, err
		}

		if w.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}
		w.URL = string(url)
		w.LoadURL = string(loadURL)
		w.ResponseTimestamp = responseTime.UnixNano()
		log.Info().Time("url_request_time", urlRequestTime).Msg("url request time...")
		w.URLRequestTimestamp = urlRequestTime.UnixNano()
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

	if webData == nil || webData.Address == nil {
		return 0, am.ErrEmptyAddress
	}

	logger := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).
		Int("GroupID", webData.Address.GroupID).
		Int64("AddressID", webData.Address.AddressID).Logger()

	if err := s.addSnapshots(ctx, userContext, logger, webData); err != nil {
		logger.Warn().Err(err).Msg("failed to insert snapshot, serialized dom and detected tech")
	}

	logger.Info().Int64("url_request_timestamp", webData.URLRequestTimestamp).Msg("adding responses for webData")
	orgID, err := s.addResponses(ctx, userContext, logger, webData)
	if err != nil {
		return 0, err
	}
	return orgID, nil
}

func (s *Service) addSnapshots(ctx context.Context, userContext am.UserContext, logger zerolog.Logger, webData *am.WebData) error {
	var err error
	var tx *pgx.Tx

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	var snapshotID int64
	oid := webData.Address.OrgID
	gid := webData.Address.GroupID
	err = tx.QueryRowEx(ctx, "insertSnapshot", &pgx.QueryExOptions{}, oid, gid, webData.AddressHash, webData.HostAddress,
		webData.IPAddress, webData.Scheme, webData.ResponsePort, webData.RequestedPort, webData.URL, time.Unix(0, webData.ResponseTimestamp),
		webData.SerializedDOMHash, webData.SerializedDOMLink, webData.SnapshotLink, webData.LoadURL, time.Unix(0, webData.URLRequestTimestamp)).Scan(&snapshotID)

	if err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return v
		}
		return err
	}

	if webData.DetectedTech != nil {
		for techName, data := range webData.DetectedTech {
			log.Info().Msgf("adding webtech %#v", data)
			_, err := tx.ExecEx(ctx, "insertWebTech", &pgx.QueryExOptions{}, snapshotID, oid, gid, data.Matched, data.Location, data.Version, techName)
			if err != nil {
				if v, ok := err.(pgx.PgError); ok {
					log.Error().Err(v).Msg("failed to insert web tech")
					return v
				}
			}
		}
	}

	return tx.Commit()
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

		responseRows = append(responseRows, []interface{}{int32(oid), int32(gid), r.AddressHash, time.Unix(0, webData.URLRequestTimestamp), time.Unix(0, r.ResponseTimestamp), r.IsDocument, r.Scheme, r.IPAddress,
			r.HostAddress, webData.IPAddress, webData.HostAddress, responsePort, requestedPort, r.URL, r.Headers, r.Status, r.StatusText, r.MimeType, r.RawBodyHash, r.RawBodyLink,
		})

		if r.WebCertificate != nil {
			c := r.WebCertificate
			certificateRows = append(certificateRows, []interface{}{int32(oid), int32(gid), time.Unix(0, r.ResponseTimestamp), webData.AddressHash, r.HostAddress, r.IPAddress, responsePort,
				c.Protocol, c.KeyExchange, c.KeyExchangeGroup, c.Cipher, c.Mac, c.CertificateValue, c.SubjectName, c.SanList, c.Issuer, c.ValidFrom, c.ValidTo, c.CertificateTransparencyCompliance})
		}
	}

	return responseRows, certificateRows
}
