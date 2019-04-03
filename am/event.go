package am

import "context"

const (
	RNEventService  = "lrn:service:eventservice:feature:events"
	EventServiceKey = "eventservice"
)

var (
	EventInitialGroupComplete int32 = 1
	EventMaxHostPricing       int32 = 2
	EventNewHost              int32 = 10
	EventNewRecord            int32 = 11
	EventNewWebsite           int32 = 100
	EventWebHTMLUpdated       int32 = 101
	EventWebTechChanged       int32 = 102
	EventWebJSChanged         int32 = 103
	EventCertExpiring         int32 = 150
	EventCertExpired          int32 = 151
	EventAXFR                 int32 = 200
	EventNSEC                 int32 = 201
)
var EventTypes = map[int32]string{
	1:   "initial scan group analysis completed",
	2:   "maximum number of hostnames reached for pricing plan",
	10:  "new hostname",
	11:  "new record",
	100: "new website detected",
	101: "website's html updated",
	102: "website's technology changed",
	103: "website's javascript changed",
	150: "certificate expiring",
	151: "certificate expired",
	200: "dns server exposing records via zone transfer",
	201: "dns server exposing records via NSEC walking",
}

type Event struct {
	NotificationID int64    `json:"notification_id"`
	OrgID          int      `json:"org_id"`
	GroupID        int      `json:"group_id"`
	TypeID         int32    `json:"type_id"`
	EventTimestamp int64    `json:"event_timestamp"`
	Data           []string `json:"data"`
	Read           bool     `json:"read"`
}

type EventSubscriptions struct {
	TypeID              int32 `json:"type_id"`
	SubscribedTimestamp int64 `json:"subscribed_since"`
	Subscribed          bool  `json:"subscribed"`
}

type UserEventSettings struct {
	WeeklyReportSendDay int32                 `json:"weekly_report_day"`
	ShouldWeeklyEmail   bool                  `json:"should_email_weekly"`
	DailyReportSendHour int32                 `json:"daily_report_hour"`
	ShouldDailyEmail    bool                  `json:"should_email_daily"`
	UserTimezone        string                `json:"user_timezone"`
	Subscriptions       []*EventSubscriptions `json:"subscriptions"`
}

type EventFilter struct {
	Start   int64       `json:"start"`
	Limit   int32       `json:"limit"`
	Filters *FilterType `json:"filter"`
}

// EventService handles adding events and returning them to users.
type EventService interface {
	Init(config []byte) error
	// Get events
	Get(ctx context.Context, userContext UserContext, filter *EventFilter) ([]*Event, error)
	// GetSettings user settings
	GetSettings(ctx context.Context, userContext UserContext) (*UserEventSettings, error)
	// MarkRead events
	MarkRead(ctx context.Context, userContext UserContext, notificationIDs []int64) error
	// Add events (system only?)
	Add(ctx context.Context, userContext UserContext, events []*Event) error
	// UpdateSettings for user
	UpdateSettings(ctx context.Context, userContext UserContext, settings *UserEventSettings) error
	// NotifyComplete that a scan group has completed
	NotifyComplete(ctx context.Context, userContext UserContext, startTime int64, groupID int) error
}
