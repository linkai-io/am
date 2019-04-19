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
}

// New returns an empty Service
func New(authorizer auth.Authorizer, scanGroupClient am.ScanGroupService, addressClient am.AddressService) *Service {
	return &Service{
		authorizer:      authorizer,
		scanGroupClient: scanGroupClient,
		addressClient:   addressClient,
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

func (s *Service) Create(ctx context.Context, userContext am.UserContext, config *am.CustomWebFlowConfig) (webFlowID int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "create") {
		return 0, am.ErrUserNotAuthorized
	}
	var group *am.ScanGroup

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "CustomWebFlowService.Create").
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Creating CustomWebFlow")

	if _, group, err = s.scanGroupClient.Get(ctx, userContext, config.GroupID); err != nil {
		return 0, err
	}

	if group.OrgID != userContext.GetOrgID() {
		return 0, am.ErrOrgIDMismatch
	}

	// creates and sets webflowid
	err = s.pool.QueryRow("createCustomWebScan", userContext.GetOrgID(), group.GroupID, config.WebFlowName, config.Configuration, time.Now(), time.Now()).Scan(&webFlowID)
	if err != nil {
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
		Str("Call", "CustomWebFlow.Delete").
		Str("TraceID", userContext.GetTraceID()).Logger()

	serviceLog.Info().Msg("Starting web flow")
	custom := &am.CustomWebFlowConfig{}
	//organization_id, scan_group_id, web_flow_id, web_flow_name, configuration, created_timestamp, modified_timestamp, deleted
	err = tx.QueryRow("getCustomWebScan", userContext.GetOrgID(), webFlowID).Scan(&custom.OrgID, &custom.GroupID, &custom.WebFlowID,
		&custom.WebFlowName, &custom.Configuration, createTime, modifyTime, &custom.Deleted)
	if err != nil {
		return 0, err
	}
	custom.CreationTime = createTime.UnixNano()
	custom.ModifiedTime = modifyTime.UnixNano()
	return 0, nil
}

func (s *Service) Stop(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	return 0, nil
}

func (s *Service) GetStatus(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, []*am.CustomWebStatus, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return 0, nil, nil
}

func (s *Service) GetResults(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, []*am.CustomWebFlowResults, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNWebData, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return 0, nil, nil
}
