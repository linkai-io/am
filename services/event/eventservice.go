package event

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var ()

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
			log.Error().Err(err).Msgf("failed to prepare %s: %s", k, v)
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

func (s *Service) Get(ctx context.Context, userContext am.UserContext, filter *am.EventFilter) ([]*am.Event, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNEventService, "read") {
		return nil, am.ErrUserNotAuthorized
	}

	var getQuery string
	var args []interface{}
	var rows *pgx.Rows
	var tx *pgx.Tx
	var err error

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	if filter.Limit > 10000 {
		return nil, am.ErrLimitTooLarge
	}

	getQuery, args, err = buildGetFilterQuery(userContext, filter)
	if err != nil {
		return nil, err
	}

	serviceLog.Info().Str("query", getQuery).Msgf("executing query %v", args)

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	rows, err = tx.Query(getQuery, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	events := make([]*am.Event, 0)
	for i := 0; rows.Next(); i++ {
		var ts time.Time
		e := &am.Event{}
		e.Data = make([]string, 0)
		if err := rows.Scan(&e.OrgID, &e.GroupID, &e.NotificationID, &e.TypeID, &ts, &e.Data); err != nil {
			return nil, err
		}
		e.EventTimestamp = ts.UnixNano()
		events = append(events, e)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *Service) GetSettings(ctx context.Context, userContext am.UserContext) (*am.UserEventSettings, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNEventService, "read") {
		return nil, am.ErrUserNotAuthorized
	}

	var rows *pgx.Rows
	var tx *pgx.Tx
	var err error

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	oid := 0
	uid := 0
	settings := &am.UserEventSettings{}
	err = tx.QueryRow("getUserSettings", userContext.GetOrgID(), userContext.GetUserID()).Scan(&oid, &uid, &settings.WeeklyReportSendDay, &settings.DailyReportSendHour, &settings.UserTimezone, &settings.ShouldWeeklyEmail, &settings.ShouldDailyEmail)
	if err != nil {
		return nil, err
	}

	if oid != userContext.GetOrgID() {
		return nil, am.ErrOrgIDMismatch
	}

	if uid != userContext.GetUserID() {
		return nil, am.ErrUserIDMismatch
	}

	rows, err = tx.Query("getUserSubscriptions", userContext.GetOrgID(), userContext.GetUserID())
	if err != nil {
		return nil, err
	}

	settings.Subscriptions = make([]*am.EventSubscriptions, 0)
	for i := 0; rows.Next(); i++ {
		sub := &am.EventSubscriptions{}
		var ts time.Time
		if err := rows.Scan(&oid, &uid, &sub.TypeID, &ts, &sub.Subscribed); err != nil {
			return nil, err
		}

		if oid != userContext.GetOrgID() {
			return nil, am.ErrOrgIDMismatch
		}

		if uid != userContext.GetUserID() {
			return nil, am.ErrUserIDMismatch
		}
		sub.SubscribedTimestamp = ts.UnixNano()
		settings.Subscriptions = append(settings.Subscriptions, sub)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return settings, nil
}

// MarkRead events
func (s *Service) MarkRead(ctx context.Context, userContext am.UserContext, notificationIDs []int64) error {
	if !s.IsAuthorized(ctx, userContext, am.RNEventService, "update") {
		return am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	ctx = serviceLog.WithContext(ctx)

	var tx *pgx.Tx
	var err error

	log.Ctx(ctx).Info().Int("notifyid_len", len(notificationIDs)).Msg("adding")

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if _, err := tx.Exec(MarkReadTempTable); err != nil {
		return err
	}

	numEvents := len(notificationIDs)

	eventRows := make([][]interface{}, numEvents)
	orgID := userContext.GetOrgID()
	userID := userContext.GetUserID()

	for i := 0; i < len(notificationIDs); i++ {
		eventRows[i] = []interface{}{orgID, userID, notificationIDs[i]}
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{MarkReadTempTableKey}, MarkReadTempTableColumns, pgx.CopyFromRows(eventRows))
	if err != nil {
		return err
	}

	if copyCount != numEvents {
		return am.ErrEventCopyCount
	}

	if _, err := tx.Exec(MarkReadTempToMarkRead); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, "failed to mark events as read")
		}
		return err
	}

	err = tx.Commit()

	return err
}

// Add events
func (s *Service) Add(ctx context.Context, userContext am.UserContext, events []*am.Event) error {
	if !s.IsAuthorized(ctx, userContext, am.RNEventService, "create") {
		return am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	ctx = serviceLog.WithContext(ctx)

	var tx *pgx.Tx
	var err error

	numEvents := len(events)
	log.Ctx(ctx).Info().Int("event_len", numEvents).Msg("adding")

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if _, err := tx.Exec(AddTempTable); err != nil {
		return err
	}

	eventRows := make([][]interface{}, numEvents)
	orgID := userContext.GetOrgID()
	//userID := userContext.GetUserID()

	for i := 0; i < numEvents; i++ {
		log.Ctx(ctx).Info().Msgf("%#v", events[i])
		eventRows[i] = []interface{}{orgID, events[i].GroupID, events[i].TypeID, time.Unix(0, events[i].EventTimestamp), events[i].Data}
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{AddTempTableKey}, AddTempTableColumns, pgx.CopyFromRows(eventRows))
	if err != nil {
		return err
	}

	if copyCount != numEvents {
		return am.ErrEventCopyCount
	}

	if _, err := tx.Exec(AddTempToAdd); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, "failed to add events")
		}
		return err
	}

	return tx.Commit()
}

// UpdateSettings for user
func (s *Service) UpdateSettings(ctx context.Context, userContext am.UserContext, settings *am.UserEventSettings) error {
	if !s.IsAuthorized(ctx, userContext, am.RNEventService, "update") {
		return am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	ctx = serviceLog.WithContext(ctx)

	var tx *pgx.Tx
	var err error

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if err = s.updateSubscriptions(ctx, userContext, tx, settings.Subscriptions); err != nil {
		return err
	}
	_, err = tx.Exec("updateUserSettings", userContext.GetOrgID(), userContext.GetUserID(), settings.WeeklyReportSendDay, settings.DailyReportSendHour, settings.UserTimezone, settings.ShouldWeeklyEmail, settings.ShouldDailyEmail)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) updateSubscriptions(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, subscriptions []*am.EventSubscriptions) error {
	numSubscriptions := len(subscriptions)
	log.Ctx(ctx).Info().Int("subscriptions", numSubscriptions).Msg("adding")
	if _, err := tx.Exec(SubscriptionsTempTable); err != nil {
		return err
	}

	subRows := make([][]interface{}, numSubscriptions)
	orgID := userContext.GetOrgID()
	userID := userContext.GetUserID()

	for i := 0; i < numSubscriptions; i++ {
		subRows[i] = []interface{}{orgID, userID, subscriptions[i].TypeID, time.Unix(0, subscriptions[i].SubscribedTimestamp), subscriptions[i].Subscribed}
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{SubscriptionsTempTableKey}, SubscriptionsTempTableColumns, pgx.CopyFromRows(subRows))
	if err != nil {
		return err
	}

	if copyCount != numSubscriptions {
		return am.ErrEventCopyCount
	}

	if _, err := tx.Exec(SubscriptionsTempToSubscriptions); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, "failed to add subscriptions")
		}
		return err
	}
	return nil
}

// NotifyComplete that a scan group has completed
func (s *Service) NotifyComplete(ctx context.Context, userContext am.UserContext, startTime int64, groupID int) error {
	if !s.IsAuthorized(ctx, userContext, am.RNEventService, "update") {
		return am.ErrUserNotAuthorized
	}
	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("call", "event.NotifyComplete").
		Str("TraceID", userContext.GetTraceID()).Logger()

	ctx = serviceLog.WithContext(ctx)

	events := make([]*am.Event, 0)
	tx, err := s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback() // safe to call as no-op on success
	newHosts, err := s.newHostnames(ctx, userContext, tx, startTime, groupID)
	if err != nil {
		serviceLog.Error().Err(err).Msg("failed to gather new hosts events")
	} else if newHosts != nil {
		events = append(events, newHosts)
	}

	// new websites
	newWebsites, err := s.newWebsites(ctx, userContext, tx, startTime, groupID)
	if err != nil {
		serviceLog.Error().Err(err).Msg("failed to gather new websites events")
	} else if newWebsites != nil {
		events = append(events, newWebsites)
	}
	// diff websites

	// test web tech

	// check certificates
	expiringCerts, err := s.expiringCerts(ctx, userContext, tx, startTime, groupID)
	if err != nil {
		serviceLog.Error().Err(err).Msg("failed to gather new certificate expiration events")
	} else if expiringCerts != nil {
		events = append(events, expiringCerts)
	}
	if err := tx.Commit(); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, "failed to notify complete")
		}
		return err
	}

	return s.Add(ctx, userContext, events)
}

func (s *Service) expiringCerts(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, startTime int64, groupID int) (*am.Event, error) {
	oid := userContext.GetOrgID()
	// check new hostnames
	rows, err := tx.Query("checkCertExpiration", oid, groupID, time.Unix(0, startTime))
	if err != nil {
		return nil, err
	}

	certs := make([]string, 0)
	for i := 0; rows.Next(); i++ {
		var subjectName string
		var port int
		var validTo int64
		if err := rows.Scan(&subjectName, &port, &validTo); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to scan new certificate expiring event")
			continue
		}
		validTime := strconv.FormatInt(validTo, 10)

		certs = append(certs, subjectName)
		certs = append(certs, fmt.Sprintf("%d", port))
		certs = append(certs, validTime)
	}
	// no new certs this round
	if len(certs) == 0 {
		return nil, nil
	}

	e := &am.Event{
		OrgID:          oid,
		GroupID:        groupID,
		TypeID:         am.EventCertExpiring,
		EventTimestamp: time.Now().UnixNano(),
		Data:           certs,
	}
	return e, nil
}

func (s *Service) newWebsites(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, startTime int64, groupID int) (*am.Event, error) {
	oid := userContext.GetOrgID()
	// check new hostnames
	rows, err := tx.Query("newWebsites", oid, groupID, time.Unix(0, startTime), oid, groupID)
	if err != nil {
		return nil, err
	}
	urlPorts := make([]string, 0)
	for i := 0; rows.Next(); i++ {
		var url []byte
		var port int
		if err := rows.Scan(&url, &port); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to scan new website event")
			continue
		}
		urlPorts = append(urlPorts, string(url))
		urlPorts = append(urlPorts, fmt.Sprintf("%d", port))
	}
	// no new urls this round
	if len(urlPorts) == 0 {
		return nil, nil
	}

	e := &am.Event{
		OrgID:          oid,
		GroupID:        groupID,
		TypeID:         am.EventNewWebsite,
		EventTimestamp: time.Now().UnixNano(),
		Data:           urlPorts,
	}
	return e, nil
}

func (s *Service) newHostnames(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, startTime int64, groupID int) (*am.Event, error) {
	oid := userContext.GetOrgID()
	// check new hostnames
	rows, err := tx.Query("newHostnames", oid, groupID, time.Unix(0, startTime), oid, groupID)
	if err != nil {
		return nil, err
	}
	hosts := make([]string, 0)
	for i := 0; rows.Next(); i++ {

		var host string
		if err := rows.Scan(&host); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to scan new hostname event")
			continue
		}
		hosts = append(hosts, host)

	}
	// no new hosts this round
	if len(hosts) == 0 {
		return nil, nil
	}

	e := &am.Event{
		OrgID:          oid,
		GroupID:        groupID,
		TypeID:         am.EventNewHost,
		EventTimestamp: time.Now().UnixNano(),
		Data:           hosts,
	}
	return e, nil
}
