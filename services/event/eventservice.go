package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/linkai-io/am/pkg/webhooks"

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
	webhooks   webhooks.Webhooker
}

// New returns an empty Service
func New(authorizer auth.Authorizer, hooks webhooks.Webhooker) *Service {
	return &Service{authorizer: authorizer, webhooks: hooks}
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
	var tx *pgx.Tx
	var rows *pgx.Rows
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
	serviceLog.Info().Msg("reading rows")

	events := make([]*am.Event, 0)
	for i := 0; rows.Next(); i++ {
		var ts time.Time
		e := &am.Event{}
		e.Data = make([]string, 0)
		if err := rows.Scan(&e.OrgID, &e.GroupID, &e.NotificationID, &e.TypeID, &ts, &e.Data, &e.JSONData); err != nil {
			return nil, err
		}
		e.EventTimestamp = ts.UnixNano()
		events = append(events, e)
	}
	if err := tx.Commit(); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return nil, v
		}
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
		if v, ok := err.(pgx.PgError); ok {
			return nil, v
		}
		return nil, err
	}
	return settings, nil
}

func (s *Service) GetWebhooks(ctx context.Context, userContext am.UserContext) ([]*am.WebhookEventSettings, error) {
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

	hooks := make([]*am.WebhookEventSettings, 0)
	rows, err = tx.Query("getWebhooksForUser", userContext.GetOrgID())
	if err != nil {
		return nil, err
	}

	for i := 0; rows.Next(); i++ {
		w := &am.WebhookEventSettings{}
		if err := rows.Scan(&w.WebhookID, &w.OrgID, &w.GroupID, &w.Name, &w.Events, &w.Enabled,
			&w.Version, &w.URL, &w.Type, &w.CurrentKey, &w.PreviousKey, &w.Deleted, &w.ScanGroupName); err != nil {
			return nil, err
		}

		if w.OrgID != int32(userContext.GetOrgID()) {
			return nil, am.ErrOrgIDMismatch
		}
		hooks = append(hooks, w)
	}

	if err := tx.Commit(); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return nil, v
		}
		return nil, err
	}
	return hooks, nil
}

// UpdateWebhooks or insert new (uniqueness on org,groupid,name)
func (s *Service) UpdateWebhooks(ctx context.Context, userContext am.UserContext, webhook *am.WebhookEventSettings) (err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNEventService, "create") {
		return am.ErrUserNotAuthorized
	}

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "event.UpdateWebhooks").
		Str("TraceID", userContext.GetTraceID()).Logger()

	ctx = serviceLog.WithContext(ctx)
	var tx *pgx.Tx

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if webhook.Deleted == true {
		deletedName := fmt.Sprintf("%s_%d", webhook.Name, time.Now().UnixNano())
		if _, err = tx.Exec("deleteWebhook", deletedName, userContext.GetOrgID(), webhook.GroupID, webhook.Name); err != nil {
			return err
		}
	} else {
		if _, err = tx.Exec("updateWebhook", userContext.GetOrgID(), webhook.GroupID, webhook.Name, webhook.Events,
			webhook.Enabled, webhook.Version, webhook.URL, webhook.Type, webhook.CurrentKey, webhook.PreviousKey); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) GetWebhookEvents(ctx context.Context, userContext am.UserContext) ([]*am.WebhookEvent, error) {
	if !s.IsAuthorized(ctx, userContext, am.RNEventService, "read") {
		return nil, am.ErrUserNotAuthorized
	}

	var tx *pgx.Tx
	var rows *pgx.Rows
	var err error

	serviceLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("OrgID", userContext.GetOrgID()).
		Str("Call", "event.GetWebhookEvents").
		Str("TraceID", userContext.GetTraceID()).Logger()
	ctx = serviceLog.WithContext(ctx)

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	rows, err = tx.Query("getLastWebhookEvents", userContext.GetOrgID())
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	serviceLog.Info().Msg("reading rows")

	events := make([]*am.WebhookEvent, 0)
	for i := 0; rows.Next(); i++ {
		var ts time.Time
		e := &am.WebhookEvent{}
		if err := rows.Scan(&e.OrgID, &e.GroupID, &e.NotificationID, &e.WebhookID, &e.TypeID, &ts, &e.LastAttemptStatus); err != nil {
			return nil, err
		}
		e.LastAttemptTimestamp = ts.UnixNano()
		events = append(events, e)
	}

	if err := tx.Commit(); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return nil, v
		}
		return nil, err
	}
	return events, nil
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
	if numEvents == 0 {
		return nil
	}
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

	for i := 0; i < numEvents; i++ {
		log.Ctx(ctx).Info().Msgf("%#v", events[i])
		eventRows[i] = []interface{}{orgID, events[i].GroupID, events[i].TypeID, time.Unix(0, events[i].EventTimestamp), events[i].Data, events[i].JSONData}
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{AddTempTableKey}, AddTempTableColumns, pgx.CopyFromRows(eventRows))
	if err != nil {
		return err
	}

	if copyCount != numEvents {
		return am.ErrEventCopyCount
	}

	rows, err := tx.Query(AddTempToAdd)
	if err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, "failed to add events")
		}
		return err
	}
	log.Ctx(ctx).Info().Msgf("got rows: %v\n", rows)

	webhookEvents := make([]*am.Event, 0)
	for rows.Next() {
		evt := &am.Event{}
		var evtTime time.Time
		if err := rows.Scan(&evt.NotificationID, &evt.OrgID, &evt.GroupID, &evt.TypeID, &evtTime, &evt.JSONData); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("failed to read back event from insert with notification id")
			continue
		}
		evt.EventTimestamp = evtTime.UnixNano()
		webhookEvents = append(webhookEvents, evt)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	allHooks, err := s.getOrgWebhooks(ctx, userContext, webhookEvents)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("unable to retrieve webhooks")
	}

	if len(allHooks) > 0 {
		return s.sendWebhookEvents(ctx, userContext, allHooks)
	}
	return nil
}

func (s *Service) sendWebhookEvents(ctx context.Context, userContext am.UserContext, hooks []*webhooks.Data) error {
	responses := make([]*webhooks.DataResponse, 0)
	for _, hook := range hooks {

		resp, err := s.webhooks.Send(ctx, hook)
		if err != nil {
			continue
		}
		resp.Data = hook
		responses = append(responses, resp)
	}

	if len(responses) == 0 {
		log.Ctx(ctx).Warn().Msg("no responses from send webhooks")
		return nil
	}

	tx, err := s.pool.BeginEx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if _, err := tx.Exec(WebhookTempTable); err != nil {
		return err
	}

	eventRows := make([][]interface{}, 0)
	orgID := userContext.GetOrgID()
	for _, resp := range responses {
		for _, evt := range resp.Data.Event {
			eventRows = append(eventRows, []interface{}{orgID, resp.Data.Settings.GroupID, evt.NotificationID, resp.Data.Settings.WebhookID, evt.TypeID, time.Unix(0, resp.DeliveredTime), resp.StatusCode})
		}
	}

	copyCount, err := tx.CopyFrom(pgx.Identifier{WebhookTempTableKey}, WebhookTempTableColumns, pgx.CopyFromRows(eventRows))
	if err != nil {
		return err
	}

	if copyCount != len(eventRows) {
		return am.ErrEventCopyCount
	}

	return tx.Commit()
}

// getOrgWebhooks scans the table with webhook settings and only return webhooks that match defined hooks for this scangroup + actual event types
func (s *Service) getOrgWebhooks(ctx context.Context, userContext am.UserContext, events []*am.Event) ([]*webhooks.Data, error) {
	allHooks := make([]*webhooks.Data, 0)

	// assumes len(events) > 0 and all events are from same group (which they should be)
	groupID := events[0].GroupID

	rows, err := s.pool.QueryEx(ctx, "getOrgWebhooks", &pgx.QueryExOptions{}, userContext.GetOrgID(), groupID)
	if err != nil {
		return allHooks, err
	}

	for rows.Next() {
		w := &am.WebhookEventSettings{}
		if err := rows.Scan(&w.WebhookID, &w.OrgID, &w.GroupID, &w.Name, &w.Events, &w.Enabled,
			&w.Version, &w.URL, &w.Type, &w.CurrentKey, &w.PreviousKey, &w.Deleted, &w.ScanGroupName); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("failed to scan webhook row")
			continue
		}

		evt := &webhooks.Data{
			Settings: w,
			Event:    make([]*am.Event, 0),
		}

		// iterate over events, and make sure this webhook is subscribed to the event type
		// if it is, add it to the events specific for this webhook.
		for _, e := range events {
			for _, subscribed := range w.Events {
				if e.TypeID == subscribed {
					evt.Event = append(evt.Event, e)
				}
			}
		}

		// if no webhooks match the event types, don't bother adding to our list of webhooks
		if len(evt.Event) == 0 {
			continue
		}
		allHooks = append(allHooks, evt)
	}
	return allHooks, nil
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
	newTechnologies, err := s.newTech(ctx, userContext, tx, startTime, groupID)
	if err != nil {
		serviceLog.Error().Err(err).Msg("failed to gather new tech events")
	} else if newTechnologies != nil {
		events = append(events, newTechnologies)
	}

	// check certificates
	expiringCerts, err := s.expiringCerts(ctx, userContext, tx, startTime, groupID)
	serviceLog.Info().Msgf("expiring certs: %#v\n", expiringCerts)
	if err != nil {
		serviceLog.Error().Err(err).Msg("failed to gather new certificate expiration events")
	} else if expiringCerts != nil {
		events = append(events, expiringCerts)
	}

	// port changes
	portChanges, err := s.portChanges(ctx, userContext, tx, startTime, groupID)
	serviceLog.Info().Msgf("port changes: %#v\n", portChanges)
	if err != nil {
		serviceLog.Error().Err(err).Msg("failed to gather port change events")
	} else if portChanges != nil {
		events = append(events, portChanges...)
	}

	if err := tx.Commit(); err != nil {
		if v, ok := err.(pgx.PgError); ok {
			return errors.Wrap(v, "failed to notify complete")
		}
		return err
	}

	return s.Add(ctx, userContext, events)
}

func (s *Service) portChanges(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, startTime int64, groupID int) ([]*am.Event, error) {
	oid := userContext.GetOrgID()
	rows, err := tx.Query("checkPortChanges", oid, groupID, time.Unix(0, startTime))
	if err != nil {
		return nil, err
	}

	openPorts := make([]*am.EventNewOpenPort, 0)
	closedPorts := make([]*am.EventClosedPort, 0)
	for i := 0; rows.Next(); i++ {
		var ports am.Ports
		var host string
		var scanTime time.Time
		var prevTime time.Time

		if err := rows.Scan(&host, &ports, &scanTime, &prevTime); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to scan port change event")
			continue
		}
		open, closed, change := ports.TCPChanges()
		if change {
			log.Ctx(ctx).Info().Msg("DETECTED CHANGE")
			if len(open) > 0 {
				openPorts = append(openPorts, &am.EventNewOpenPort{
					Host:       host,
					CurrentIP:  ports.Current.IPAddress,
					PreviousIP: ports.Previous.IPAddress,
					OpenPorts:  open,
				})
			}
			if len(closed) > 0 {
				closedPorts = append(closedPorts, &am.EventClosedPort{
					Host:        host,
					CurrentIP:   ports.Current.IPAddress,
					PreviousIP:  ports.Previous.IPAddress,
					ClosedPorts: closed,
				})
			}
		}

	}
	events := make([]*am.Event, 0)
	if len(openPorts) != 0 {
		m, err := json.Marshal(openPorts)
		if err != nil {
			return nil, err
		}

		events = append(events, &am.Event{
			OrgID:          oid,
			GroupID:        groupID,
			TypeID:         am.EventNewOpenPortID,
			EventTimestamp: time.Now().UnixNano(),
			JSONData:       string(m),
		})
	}

	if len(closedPorts) != 0 {
		m, err := json.Marshal(closedPorts)
		if err != nil {
			return nil, err
		}
		events = append(events, &am.Event{
			OrgID:          oid,
			GroupID:        groupID,
			TypeID:         am.EventClosedPortID,
			EventTimestamp: time.Now().UnixNano(),
			JSONData:       string(m),
		})
	}
	if len(events) == 0 {
		return nil, nil
	}

	return events, nil

}

func (s *Service) expiringCerts(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, startTime int64, groupID int) (*am.Event, error) {
	oid := userContext.GetOrgID()
	// check new hostnames
	rows, err := tx.Query("checkCertExpiration", oid, groupID, time.Unix(0, startTime))
	if err != nil {
		return nil, err
	}

	certs := make([]*am.EventCertExpiring, 0)
	for i := 0; rows.Next(); i++ {
		var subjectName string
		var port int
		var validTo int64
		if err := rows.Scan(&subjectName, &port, &validTo); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to scan new certificate expiring event")
			continue
		}

		certs = append(certs, &am.EventCertExpiring{
			SubjectName: subjectName,
			Port:        port,
			ValidTo:     validTo,
		})

	}
	// no new certs this round
	if len(certs) == 0 {
		return nil, nil
	}

	m, err := json.Marshal(certs)
	if err != nil {
		return nil, err
	}

	e := &am.Event{
		OrgID:          oid,
		GroupID:        groupID,
		TypeID:         am.EventCertExpiringID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
	}
	return e, nil
}

func (s *Service) newWebsites(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, startTime int64, groupID int) (*am.Event, error) {
	oid := userContext.GetOrgID()
	// check new hostnames
	rows, err := tx.Query("newWebsites", oid, groupID, time.Unix(0, startTime))
	if err != nil {
		return nil, err
	}
	newSites := make([]*am.EventNewWebsite, 0)
	for i := 0; rows.Next(); i++ {
		var loadURL []byte
		var url []byte
		var port int
		if err := rows.Scan(&loadURL, &url, &port); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to scan new website event")
			continue
		}
		newSites = append(newSites, &am.EventNewWebsite{
			LoadURL: string(loadURL),
			URL:     string(url),
			Port:    port,
		})
	}
	// no new urls this round
	if len(newSites) == 0 {
		return nil, nil
	}

	m, err := json.Marshal(newSites)
	if err != nil {
		return nil, err
	}

	e := &am.Event{
		OrgID:          oid,
		GroupID:        groupID,
		TypeID:         am.EventNewWebsiteID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
	}
	return e, nil
}

func (s *Service) newTech(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, startTime int64, groupID int) (*am.Event, error) {
	oid := userContext.GetOrgID()
	// check new hostnames
	rows, err := tx.Query("newTechnologies", oid, groupID, time.Unix(0, startTime))
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error getting new tech")
		return nil, err
	}

	techPorts := make([]*am.EventNewWebTech, 0)
	for i := 0; rows.Next(); i++ {
		var loadURL []byte
		var url []byte
		var port int
		var techName string
		var version string
		if err := rows.Scan(&loadURL, &url, &port, &techName, &version); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to scan new technology event")
			continue
		}
		log.Ctx(ctx).Info().Msgf("%s %d %s %s\n", string(url), port, techName, version)

		techPorts = append(techPorts, &am.EventNewWebTech{
			LoadURL:  string(loadURL),
			URL:      string(url),
			Port:     port,
			TechName: techName,
			Version:  version,
		})
	}
	// no new technologies this round
	if len(techPorts) == 0 {
		return nil, nil
	}

	m, err := json.Marshal(techPorts)
	if err != nil {
		return nil, err
	}

	e := &am.Event{
		OrgID:          oid,
		GroupID:        groupID,
		TypeID:         am.EventNewWebTechID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
	}
	return e, nil
}

func (s *Service) newHostnames(ctx context.Context, userContext am.UserContext, tx *pgx.Tx, startTime int64, groupID int) (*am.Event, error) {
	oid := userContext.GetOrgID()
	// check new hostnames
	rows, err := tx.Query("newHostnames", oid, groupID, time.Unix(0, startTime))
	if err != nil {
		return nil, err
	}
	hosts := make([]*am.EventNewHost, 0)
	for i := 0; rows.Next(); i++ {

		var host string
		if err := rows.Scan(&host); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to scan new hostname event")
			continue
		}
		hosts = append(hosts, &am.EventNewHost{Host: host})

	}
	// no new hosts this round
	if len(hosts) == 0 {
		return nil, nil
	}

	m, err := json.Marshal(hosts)
	if err != nil {
		return nil, err
	}

	e := &am.Event{
		OrgID:          oid,
		GroupID:        groupID,
		TypeID:         am.EventNewHostID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
	}
	return e, nil
}
