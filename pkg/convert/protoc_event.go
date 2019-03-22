package convert

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/protocservices/prototypes"
)

func DomainToEvent(in *am.Event) *prototypes.EventData {
	data := make(map[string]*prototypes.StringEvent, 0)
	if in.Data != nil {
		for k, v := range in.Data {
			data[k] = &prototypes.StringEvent{Value: v}
		}
	}

	return &prototypes.EventData{
		EventID:        in.EventID,
		OrgID:          int32(in.OrgID),
		GroupID:        int32(in.GroupID),
		TypeID:         in.TypeID,
		EventTimestamp: in.EventTimestamp,
		Data:           data,
		Read:           false,
	}
}

func EventToDomain(in *prototypes.EventData) *am.Event {
	data := make(map[string][]string, 0)
	if in.Data != nil {
		for k, v := range in.Data {
			data[k] = v.Value
		}
	}

	return &am.Event{
		OrgID:          int(in.OrgID),
		GroupID:        int(in.GroupID),
		EventID:        in.EventID,
		TypeID:         in.TypeID,
		EventTimestamp: in.EventTimestamp,
		Data:           data,
		Read:           in.Read,
	}
}

func DomainToEventSubscriptions(in *am.EventSubscriptions) *prototypes.EventSubscriptions {
	return &prototypes.EventSubscriptions{
		TypeID:              in.TypeID,
		SubscribedTimestamp: in.SubscribedTimestamp,
	}
}

func EventSubscriptionsToDomain(in *prototypes.EventSubscriptions) *am.EventSubscriptions {
	return &am.EventSubscriptions{
		TypeID:              in.TypeID,
		SubscribedTimestamp: in.SubscribedTimestamp,
	}
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
		DailyReportSendHour: in.DailyReportSendHour,
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
		DailyReportSendHour: in.DailyReportSendHour,
		UserTimezone:        in.UserTimezone,
		Subscriptions:       subs,
	}
}

func DomainToUserEvents(in *am.UserEvents) *prototypes.UserEvents {
	events := make([]*prototypes.EventData, 0)
	if in.Events != nil {
		for _, event := range in.Events {
			events = append(events, DomainToEvent(event))
		}
	}
	return &prototypes.UserEvents{
		OrgID:    int32(in.OrgID),
		UserID:   int32(in.UserID),
		Settings: DomainToUserEventSettings(in.Settings),
		Events:   events,
	}
}

func UserEventsToDomain(in *prototypes.UserEvents) *am.UserEvents {
	events := make([]*am.Event, 0)
	if in.Events != nil {
		for _, event := range in.Events {
			events = append(events, EventToDomain(event))
		}
	}
	return &am.UserEvents{
		OrgID:    int(in.OrgID),
		UserID:   int(in.UserID),
		Settings: UserEventSettingsToDomain(in.Settings),
		Events:   events,
	}
}

func DomainToEventFilter(in *am.EventFilter) *prototypes.EventFilter {
	return &prototypes.EventFilter{
		Filters: DomainToFilterTypes(in.Filters),
	}
}

func EventFilterToDomain(in *prototypes.EventFilter) *am.EventFilter {
	return &am.EventFilter{
		Filters: FilterTypesToDomain(in.Filters),
	}
}
