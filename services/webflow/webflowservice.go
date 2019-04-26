package webflow

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
	"github.com/pkg/errors"
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
	pool            *pgx.ConnPool
	config          *pgx.ConnPoolConfig
	authorizer      auth.Authorizer
	scanGroupClient am.ScanGroupService
	addressClient   am.AddressService
	requester       WebFlowRequester
}

// New returns an empty Service
func New(authorizer auth.Authorizer, scanGroupClient am.ScanGroupService, addressClient am.AddressService, requester WebFlowRequester) *Service {
	return &Service{
		authorizer:      authorizer,
		scanGroupClient: scanGroupClient,
		addressClient:   addressClient,
		requester:       requester,
	}
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
			return errors.Wrap(err, "key: "+k)
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

func (s *Service) Create(ctx context.Context, userContext am.UserContext, config *am.CustomWebFlowConfig) (webFlowID int32, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "create") {
		return 0, am.ErrUserNotAuthorized
	}

	var group *am.ScanGroup
	var tx *pgx.Tx

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "CustomWebFlowService.Create").
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Creating CustomWebFlow")

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if _, group, err = s.scanGroupClient.Get(ctx, userContext, config.GroupID); err != nil {
		return 0, err
	}

	if group.OrgID != userContext.GetOrgID() {
		return 0, am.ErrOrgIDMismatch
	}

	// creates and sets webflowid
	err = tx.QueryRow("createCustomWebScan", userContext.GetOrgID(), group.GroupID, config.WebFlowName, config.Configuration, time.Now(), time.Now()).Scan(&webFlowID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return webFlowID, err
}

// TODO: implement later
func (s *Service) Update(ctx context.Context, userContext am.UserContext, config *am.CustomWebFlowConfig) (int, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	return 0, nil
}

func (s *Service) Delete(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "delete") {
		return 0, am.ErrUserNotAuthorized
	}

	var tx *pgx.Tx
	var name string
	var err error

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int32("WebFlowID", webFlowID).
		Str("Call", "CustomWebFlow.Delete").
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Deleting web flow")

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	// get the current group name so we can change it on delete.
	err = tx.QueryRow("customWebScanName", userContext.GetOrgID(), webFlowID).Scan(&webFlowID, &name)
	if err != nil {
		return 0, err
	}

	// ensure room for timestamp
	if len(name) > 200 {
		name = name[:200]
	}

	name = fmt.Sprintf("%s_%d\n", name, time.Now().UnixNano())

	_, err = tx.Exec("deleteCustomWebScan", name, userContext.GetOrgID(), webFlowID)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	return userContext.GetOrgID(), err
}

func (s *Service) Start(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	var tx *pgx.Tx
	var err error
	var createTime time.Time
	var modifyTime time.Time

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int32("WebFlowID", webFlowID).
		Str("Call", "CustomWebFlow.Start").
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Starting web flow")
	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	_, status, err := s.getStatus(ctx, userContext, tx, webFlowID)
	if err != nil && err != pgx.ErrNoRows {
		return 0, err
	}

	if status != nil && (status.WebFlowStatus == am.WebFlowStatusRunning) {
		return 0, errors.New("this web flow is already running")
	}

	custom := &am.CustomWebFlowConfig{}
	custom.Configuration = &am.CustomRequestConfig{}
	//organization_id, scan_group_id, web_flow_id, web_flow_name, configuration, created_timestamp, modified_timestamp, deleted
	err = tx.QueryRow("getCustomWebScan", userContext.GetOrgID(), webFlowID).Scan(&custom.OrgID, &custom.GroupID, &custom.WebFlowID,
		&custom.WebFlowName, custom.Configuration, &createTime, &modifyTime, &custom.Deleted)
	if err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to start web flow")
		}
		return 0, err
	}

	custom.CreationTime = createTime.UnixNano()
	custom.ModifiedTime = modifyTime.UnixNano()

	if _, err := tx.Exec("startStopCustomWeb", am.WebFlowStatusRunning, userContext.GetOrgID(), webFlowID); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to start web flow")
		}
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to commit start web flow")
		}
		return 0, err
	}

	executor := NewWebFlowExecutor(userContext, s, s.addressClient, s.scanGroupClient, s.requester)
	if err := executor.Init(); err != nil {
		return 0, errors.Wrap(err, "unable to start custom web flow")
	}

	if err := executor.Start(ctx, custom); err != nil {
		return 0, errors.Wrap(err, "unable to start custom web flow")
	}

	return 0, nil
}

func (s *Service) Stop(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	var tx *pgx.Tx
	var err error
	var createTime time.Time
	var modifyTime time.Time

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int32("WebFlowID", webFlowID).
		Str("Call", "CustomWebFlow.Start").
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Starting web flow")
	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	_, status, err := s.getStatus(ctx, userContext, tx, webFlowID)
	if err != nil {
		return 0, err
	}

	if status.WebFlowStatus == am.WebFlowStatusRunning {
		return 0, errors.New("this web flow is already running")
	}

	custom := &am.CustomWebFlowConfig{}
	//organization_id, scan_group_id, web_flow_id, web_flow_name, configuration, created_timestamp, modified_timestamp, deleted
	err = tx.QueryRow("getCustomWebScan", userContext.GetOrgID(), webFlowID).Scan(&custom.OrgID, &custom.GroupID, &custom.WebFlowID,
		&custom.WebFlowName, &custom.Configuration, createTime, modifyTime, &custom.Deleted)
	if err != nil {
		return 0, err
	}

	custom.CreationTime = createTime.UnixNano()
	custom.ModifiedTime = modifyTime.UnixNano()

	if _, err := tx.Exec("startStopCustomWeb", am.WebFlowStatusStopped, userContext.GetOrgID(), webFlowID); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to start web flow")
		}
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, errors.Wrap(v, "failed to commit start web flow")
		}
		return 0, err
	}

	return 0, nil
}

func (s *Service) UpdateStatus(ctx context.Context, userContext am.UserContext, total, inProgress, completed, webFlowID int32) error {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "update") {
		return am.ErrUserNotAuthorized
	}

	if _, err := s.pool.Exec("updateCustomWebStatus", total, inProgress, completed, userContext.GetOrgID(), webFlowID); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, "failed to update web flow status")
		}
		return err
	}
	return nil
}

func (s *Service) GetStatus(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, *am.CustomWebStatus, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}

	var tx *pgx.Tx
	var err error

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int32("WebFlowID", webFlowID).
		Str("Call", "CustomWebFlow.Start").
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Starting web flow")
	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	oid, status, err := s.getStatus(ctx, userContext, tx, webFlowID)
	if err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, nil, errors.Wrap(v, "failed to get web flow status")
		}
		return 0, nil, err
	}

	if err := tx.Commit(); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return 0, nil, errors.Wrap(v, "failed to commit on get web flow status")
		}
		return 0, nil, err
	}
	return oid, status, nil
}

func (s *Service) getStatus(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, webFlowID int32) (int, *am.CustomWebStatus, error) {
	//organization_id, scan_group_id, last_updated_timestamp, started_timestamp, finished_timestamp, web_flow_status, total, in_progress, completed
	status := &am.CustomWebStatus{}
	var lastUpdated time.Time
	var startTime time.Time
	var finishTime time.Time
	err := tx.QueryRow("getCustomWebScanStatus", userContext.GetOrgID(), webFlowID).Scan(&status.OrgID, &status.GroupID, &lastUpdated,
		&startTime, &finishTime, &status.WebFlowStatus, &status.Total, &status.InProgress, &status.Completed)
	if err != nil {
		return 0, nil, err
	}

	status.LastUpdatedTimestamp = lastUpdated.UnixNano()
	status.StartedTimestamp = startTime.UnixNano()
	status.FinishedTimestamp = finishTime.UnixNano()
	return status.OrgID, status, nil
}

func (s *Service) GetResults(ctx context.Context, userContext am.UserContext, filter *am.CustomWebFilter) (int, []*am.CustomWebFlowResults, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Int32("WebFlowID", filter.WebFlowID).
		Str("Call", "CustomWebFlow.GetResults").
		Str("TraceID", userContext.GetTraceID()).Logger()

	var rows *pgx.Rows
	var err error

	if filter.Limit > 10000 {
		return 0, nil, am.ErrLimitTooLarge
	}

	if filter.GroupID == 0 {
		return 0, nil, ErrFilterMissingGroupID
	}

	query, args, err := buildGetResultsQuery(userContext, filter)
	if err != nil {
		return 0, nil, err
	}
	serviceLog.Info().Msgf("executing query %s %#v", query, args)
	rows, err = s.pool.Query(query, args...)
	defer rows.Close()
	if err != nil {
		return 0, nil, err
	}

	results := make([]*am.CustomWebFlowResults, 0)

	for i := 0; rows.Next(); i++ {
		r := &am.CustomWebFlowResults{}
		var responseTime time.Time
		var runTime time.Time
		var url []byte
		var loadURL []byte

		if err := rows.Scan(&r.WebFlowID, &r.OrgID, &r.GroupID, &runTime, &url, &loadURL,
			&r.LoadHostAddress, &r.LoadIPAddress, &r.RequestedPort, &r.ResponsePort,
			&responseTime, &r.Result, &r.ResponseBodyHash, &r.ResponseBodyLink); err != nil {
			return 0, nil, err
		}

		if r.OrgID != userContext.GetOrgID() {
			return 0, nil, am.ErrOrgIDMismatch
		}
		r.URL = string(url)
		r.LoadURL = string(loadURL)
		r.ResponseTimestamp = responseTime.UnixNano()
		r.RunTimestamp = runTime.UnixNano()
		results = append(results, r)
	}

	return userContext.GetOrgID(), results, err

}

func (s *Service) AddResults(ctx context.Context, userContext am.UserContext, results []*am.CustomWebFlowResults) error {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "create") {
		return am.ErrUserNotAuthorized
	}

	if len(results) == 0 {
		return nil
	}

	var tx *pgx.Tx
	var err error

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	resultRows := make([][]interface{}, 0)
	for _, r := range results {
		resultRows = append(resultRows, []interface{}{r.WebFlowID, r.OrgID, r.GroupID, time.Unix(0, r.RunTimestamp), r.URL, r.LoadURL,
			r.LoadHostAddress, r.LoadIPAddress, r.RequestedPort, r.ResponsePort, time.Unix(0, r.ResponseTimestamp), r.Result, r.ResponseBodyHash,
			r.ResponseBodyLink})
	}

	if _, err := tx.Exec(AddWebFlowResultsTempTable); err != nil {
		return err
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{AddWebFlowResultsTempTableKey}, AddWebFlowResultsTempTableColumns, pgx.CopyFromRows(resultRows))
	if err != nil {
		return err
	}

	if copyCount != len(results) {
		return ErrCopyCount
	}

	if _, err := tx.Exec(AddTempToWebFlowResults); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, "failed to add web flow results")
		}
		return err
	}

	return tx.Commit()
}
