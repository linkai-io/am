package event

import (
	"context"
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
		e.Data = make(map[string][]string, 0)
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
		if err := rows.Scan(&oid, &uid, &sub.TypeID, &sub.SubscribedTimestamp); err != nil {
			return nil, err
		}

		if oid != userContext.GetOrgID() {
			return nil, am.ErrOrgIDMismatch
		}

		if uid != userContext.GetUserID() {
			return nil, am.ErrUserIDMismatch
		}

		settings.Subscriptions = append(settings.Subscriptions, sub)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return settings, nil
}

// MarkRead events
func (s *Service) MarkRead(ctx context.Context, userContext am.UserContext, eventIDs []int32) error {
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

	log.Ctx(ctx).Info().Int("eventid_len", len(eventIDs)).Msg("adding")

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if _, err := tx.Exec(MarkReadTempTable); err != nil {
		return err
	}

	numEvents := len(eventIDs)

	eventRows := make([][]interface{}, numEvents)
	orgID := userContext.GetOrgID()
	userID := userContext.GetUserID()

	for i := 0; i < len(eventIDs); i++ {
		eventRows[i] = []interface{}{orgID, userID, eventIDs[i]}
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
	return nil
}
