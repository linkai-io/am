package convert

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/protocservices/prototypes"
)

func DomainToEvent(in *am.Event) *prototypes.EventData {
	return &prototypes.EventData{
		NotificationID: in.NotificationID,
		OrgID:          int32(in.OrgID),
		GroupID:        int32(in.GroupID),
		TypeID:         in.TypeID,
		EventTimestamp: in.EventTimestamp,
		Data:           in.Data,
		JsonData:       in.JSONData,
		Read:           false,
	}
}

func EventToDomain(in *prototypes.EventData) *am.Event {
	return &am.Event{
		OrgID:          int(in.OrgID),
		GroupID:        int(in.GroupID),
		NotificationID: in.NotificationID,
		TypeID:         in.TypeID,
		EventTimestamp: in.EventTimestamp,
		Data:           in.Data,
		JSONData:       in.JsonData,
		Read:           in.Read,
	}
}

func DomainToEventSubscriptions(in *am.EventSubscriptions) *prototypes.EventSubscriptions {
	return &prototypes.EventSubscriptions{
		TypeID:              in.TypeID,
		SubscribedTimestamp: in.SubscribedTimestamp,
		Subscribed:          in.Subscribed,
	}
}

func EventSubscriptionsToDomain(in *prototypes.EventSubscriptions) *am.EventSubscriptions {
	return &am.EventSubscriptions{
		TypeID:              in.TypeID,
		SubscribedTimestamp: in.SubscribedTimestamp,
		Subscribed:          in.Subscribed,
	}
}

func DomainToWebhookEventSettings(in *am.WebhookEventSettings) *prototypes.WebhookEventSettings {
	return &prototypes.WebhookEventSettings{
		WebhookID:     in.WebhookID,
		OrgID:         in.OrgID,
		GroupID:       in.GroupID,
		ScanGroupName: in.ScanGroupName,
		Name:          in.Name,
		Events:        in.Events,
		Enabled:       in.Enabled,
		Version:       in.Version,
		URL:           in.URL,
		Type:          in.Type,
		CurrentKey:    in.CurrentKey,
		PreviousKey:   in.PreviousKey,
		Deleted:       in.Deleted,
	}
}

func DomainToWebhooksEventSettings(in []*am.WebhookEventSettings) []*prototypes.WebhookEventSettings {
	if in == nil || len(in) == 0 {
		return make([]*prototypes.WebhookEventSettings, 0)
	}
	hooks := make([]*prototypes.WebhookEventSettings, len(in))
	for i, setting := range in {
		hooks[i] = DomainToWebhookEventSettings(setting)
	}
	return hooks
}

func WebhookEventSettingsToDomain(in *prototypes.WebhookEventSettings) *am.WebhookEventSettings {
	return &am.WebhookEventSettings{
		WebhookID:     in.WebhookID,
		OrgID:         in.OrgID,
		GroupID:       in.GroupID,
		ScanGroupName: in.ScanGroupName,
		Name:          in.Name,
		Events:        in.Events,
		Enabled:       in.Enabled,
		Version:       in.Version,
		URL:           in.URL,
		Type:          in.Type,
		CurrentKey:    in.CurrentKey,
		PreviousKey:   in.PreviousKey,
		Deleted:       in.Deleted,
	}
}

func WebhooksEventSettingsToDomain(in []*prototypes.WebhookEventSettings) []*am.WebhookEventSettings {
	if in == nil || len(in) == 0 {
		return make([]*am.WebhookEventSettings, 0)
	}
	hooks := make([]*am.WebhookEventSettings, len(in))
	for i, setting := range in {
		hooks[i] = WebhookEventSettingsToDomain(setting)
	}
	return hooks
}

func DomainToWebhookEvent(in *am.WebhookEvent) *prototypes.WebhookEvent {
	return &prototypes.WebhookEvent{
		WebhookEventID:       in.WebhookEventID,
		OrgID:                in.OrgID,
		GroupID:              in.GroupID,
		NotificationID:       in.NotificationID,
		WebhookID:            in.WebhookID,
		TypeID:               in.TypeID,
		LastAttemptTimestamp: in.LastAttemptTimestamp,
		LastAttemptStatus:    in.LastAttemptStatus,
	}
}

func DomainToWebhookEvents(in []*am.WebhookEvent) []*prototypes.WebhookEvent {
	if in == nil || len(in) == 0 {
		return make([]*prototypes.WebhookEvent, 0)
	}
	events := make([]*prototypes.WebhookEvent, len(in))
	for i, event := range in {
		events[i] = DomainToWebhookEvent(event)
	}
	return events
}

func WebhookEventToDomain(in *prototypes.WebhookEvent) *am.WebhookEvent {
	return &am.WebhookEvent{
		WebhookEventID:       in.WebhookEventID,
		OrgID:                in.OrgID,
		GroupID:              in.GroupID,
		NotificationID:       in.NotificationID,
		WebhookID:            in.WebhookID,
		TypeID:               in.TypeID,
		LastAttemptTimestamp: in.LastAttemptTimestamp,
		LastAttemptStatus:    in.LastAttemptStatus,
	}
}

func WebhookEventsToDomain(in []*prototypes.WebhookEvent) []*am.WebhookEvent {
	if in == nil || len(in) == 0 {
		return make([]*am.WebhookEvent, 0)
	}
	events := make([]*am.WebhookEvent, len(in))
	for i, event := range in {
		events[i] = WebhookEventToDomain(event)
	}
	return events
}

func DomainToUserEventSettings(in *am.UserEventSettings) *prototypes.UserEventSettings {
	subs := make([]*prototypes.EventSubscriptions, 0)
	if in.Subscriptions != nil {
		for _, sub := range in.Subscriptions {
			subs = append(subs, DomainToEventSubscriptions(sub))
		}
	}
	return &prototypes.UserEventSettings{
		WeeklyReportSendDay: in.WeeklyReportSendDay,
		ShouldWeeklyEmail:   in.ShouldWeeklyEmail,
		DailyReportSendHour: in.DailyReportSendHour,
		ShouldDailyEmail:    in.ShouldDailyEmail,
		UserTimezone:        in.UserTimezone,
		Subscriptions:       subs,
	}
}

func UserEventSettingsToDomain(in *prototypes.UserEventSettings) *am.UserEventSettings {
	subs := make([]*am.EventSubscriptions, 0)
	if in.Subscriptions != nil {
		for _, sub := range in.Subscriptions {
			subs = append(subs, EventSubscriptionsToDomain(sub))
		}
	}
	return &am.UserEventSettings{
		WeeklyReportSendDay: in.WeeklyReportSendDay,
		ShouldWeeklyEmail:   in.ShouldWeeklyEmail,
		DailyReportSendHour: in.DailyReportSendHour,
		ShouldDailyEmail:    in.ShouldDailyEmail,
		UserTimezone:        in.UserTimezone,
		Subscriptions:       subs,
	}
}

func DomainToUserEvents(in []*am.Event) []*prototypes.EventData {
	events := make([]*prototypes.EventData, 0)
	if in != nil {
		for _, event := range in {
			events = append(events, DomainToEvent(event))
		}
	}
	return events
}

func EventsToDomain(in []*prototypes.EventData) []*am.Event {
	events := make([]*am.Event, 0)
	if in != nil {
		for _, event := range in {
			events = append(events, EventToDomain(event))
		}
	}
	return events
}

func DomainToEventFilter(in *am.EventFilter) *prototypes.EventFilter {
	return &prototypes.EventFilter{
		Start:   in.Start,
		Limit:   in.Limit,
		Filters: DomainToFilterTypes(in.Filters),
	}
}

func EventFilterToDomain(in *prototypes.EventFilter) *am.EventFilter {
	return &am.EventFilter{
		Start:   in.Start,
		Limit:   in.Limit,
		Filters: FilterTypesToDomain(in.Filters),
	}
}
