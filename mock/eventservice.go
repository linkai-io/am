package mock

import (
	"context"

	"github.com/linkai-io/am/am"
)

type EventService struct {
	InitFn        func(config []byte) error
	InitFnInvoked bool

	GetFn      func(ctx context.Context, userContext am.UserContext, filter *am.EventFilter) ([]*am.Event, error)
	GetInvoked bool

	GetSettingsFn      func(ctx context.Context, userContext am.UserContext) (*am.UserEventSettings, error)
	GetSettingsInvoked bool

	MarkReadFn      func(ctx context.Context, userContext am.UserContext, notificationIDs []int64) error
	MarkReadInvoked bool

	AddFn      func(ctx context.Context, userContext am.UserContext, event []*am.Event) error
	AddInvoked bool

	UpdateSettingsFn      func(ctx context.Context, userContext am.UserContext, settings *am.UserEventSettings) error
	UpdateSettingsInvoked bool

	NotifyCompleteFn      func(ctx context.Context, userContext am.UserContext, startTime int64, groupID int) error
	NotifyCompleteInvoked bool
}

func (s *EventService) Init(config []byte) error {
	return nil
}

func (s *EventService) Get(ctx context.Context, userContext am.UserContext, filter *am.EventFilter) ([]*am.Event, error) {
	s.GetInvoked = true
	return s.GetFn(ctx, userContext, filter)
}

func (s *EventService) GetSettings(ctx context.Context, userContext am.UserContext) (*am.UserEventSettings, error) {
	s.GetSettingsInvoked = true
	return s.GetSettingsFn(ctx, userContext)
}

func (s *EventService) MarkRead(ctx context.Context, userContext am.UserContext, notificationIDs []int64) error {
	s.MarkReadInvoked = true
	return s.MarkReadFn(ctx, userContext, notificationIDs)
}

func (s *EventService) Add(ctx context.Context, userContext am.UserContext, events []*am.Event) error {
	s.AddInvoked = true
	return s.AddFn(ctx, userContext, events)
}

func (s *EventService) UpdateSettings(ctx context.Context, userContext am.UserContext, settings *am.UserEventSettings) error {
	s.UpdateSettingsInvoked = true
	return s.UpdateSettingsFn(ctx, userContext, settings)
}

func (s *EventService) NotifyComplete(ctx context.Context, userContext am.UserContext, startTime int64, groupID int) error {
	s.NotifyCompleteInvoked = true
	return s.NotifyCompleteFn(ctx, userContext, startTime, groupID)
}
