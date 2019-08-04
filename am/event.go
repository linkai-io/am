package am

import "context"

const (
	RNEventService  = "lrn:service:eventservice:feature:events"
	EventServiceKey = "eventservice"
)

const (
	FilterEventGroupID = "group_id"
)

type Event struct {
	NotificationID int64    `json:"notification_id"`
	OrgID          int      `json:"org_id"`
	GroupID        int      `json:"group_id"`
	TypeID         int32    `json:"type_id"`
	EventTimestamp int64    `json:"event_timestamp"`
	Data           []string `json:"data,omitempty"`
	JSONData       string   `json:"json_data,omitempty"`
	Read           bool     `json:"read"`
}

type EventSubscriptions struct {
	TypeID              int32  `json:"type_id"`
	SubscribedTimestamp int64  `json:"subscribed_since"`
	Subscribed          bool   `json:"subscribed"`
	WebhookVersion      string `json:"webhook_version"`
	WebhookEnabled      bool   `json:"webhook_enabled"`
	WebhookURL          string `json:"webhook_url"`
	WebhookType         string `json:"webhook_type"`
}

type UserEventSettings struct {
	WeeklyReportSendDay int32                 `json:"weekly_report_day"`
	ShouldWeeklyEmail   bool                  `json:"should_weekly_email"`
	DailyReportSendHour int32                 `json:"daily_report_hour"`
	ShouldDailyEmail    bool                  `json:"should_daily_email"`
	UserTimezone        string                `json:"user_timezone"`
	Subscriptions       []*EventSubscriptions `json:"subscriptions"`
	WebhookCurrentKey   string                `json:"webhook_current_key,omitempty"`
	WebhookPreviousKey  string                `json:"webhook_previous_key,omitempty"`
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
