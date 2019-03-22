package am

import "context"

const (
	RNEventService = "lrn:service:eventservice:feature:events"
)

type Event struct {
	OrgID          int                 `json:"org_id"`
	GroupID        int                 `json:"group_id"`
	EventID        int32               `json:"event_id"`
	TypeID         int32               `json:"type_id"`
	EventTimestamp int64               `json:"event_timestamp"`
	Data           map[string][]string `json:"data"`
	Read           bool                `json:"read"`
}

type EventSubscriptions struct {
	TypeID              int32 `json:"type_id"`
	SubscribedTimestamp int64 `json:"subscribed_since"`
}

type UserEventSettings struct {
	WeeklyReportSendDay int32                 `json:"weekly_report_day"`
	DailyReportSendHour int32                 `json:"daily_report_hour"`
	UserTimezone        string                `json:"user_timezone"`
	Subscriptions       []*EventSubscriptions `json:"subscriptions"`
}

type UserEvents struct {
	OrgID    int                `json:"org_id"`
	UserID   int                `json:"user_id"`
	Settings *UserEventSettings `json:"settings"`
	Events   []*Event           `json:"events"`
}

type EventFilter struct {
	Filters *FilterType `json:"filter"`
}

// EventService handles adding events and returning them to users.
type EventService interface {
	Init(config []byte) error
	// Get events and user settings
	Get(ctx context.Context, userContext UserContext, filter *EventFilter) (*UserEvents, error)
	// MarkRead events
	MarkRead(ctx context.Context, userContext UserContext, eventIDs []int32) error
	// Add events (system only?)
	Add(ctx context.Context, userContext UserContext, event *Event) error
	// UpdateSettings for user
	UpdateSettings(ctx context.Context, userContext UserContext, settings *UserEventSettings) error
	// NotifyComplete that a scan group has completed
	NotifyComplete(ctx context.Context, userContext UserContext, groupID int) error
}
