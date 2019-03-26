package protoc

import (
	"errors"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/metrics/load"
	"github.com/linkai-io/am/protocservices/event"
	context "golang.org/x/net/context"
)

var (
	ErrOrgIDNonMatch      = errors.New("error organization id's did not match")
	ErrMissingUserContext = errors.New("error request was missing user context")
	ErrNilUserContext     = errors.New("error empty user context")
)

type EventProtocService struct {
	es       am.EventService
	reporter *load.RateReporter
}

func New(implementation am.EventService, reporter *load.RateReporter) *EventProtocService {
	return &EventProtocService{es: implementation, reporter: reporter}
}

func (s *EventProtocService) Get(ctx context.Context, in *event.GetRequest) (*event.GetResponse, error) {
	s.reporter.Increment(1)
	events, err := s.es.Get(ctx, convert.UserContextToDomain(in.UserContext), convert.EventFilterToDomain(in.Filter))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &event.GetResponse{Events: convert.DomainToUserEvents(events)}, nil
}

func (s *EventProtocService) GetSettings(ctx context.Context, in *event.GetSettingsRequest) (*event.GetSettingsResponse, error) {
	s.reporter.Increment(1)
	settings, err := s.es.GetSettings(ctx, convert.UserContextToDomain(in.UserContext))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &event.GetSettingsResponse{Settings: convert.DomainToUserEventSettings(settings)}, nil
}

func (s *EventProtocService) MarkRead(ctx context.Context, in *event.MarkReadRequest) (*event.MarkReadResponse, error) {
	s.reporter.Increment(1)
	err := s.es.MarkRead(ctx, convert.UserContextToDomain(in.UserContext), in.NotificationIDs)
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &event.MarkReadResponse{}, nil
}

func (s *EventProtocService) Add(ctx context.Context, in *event.AddRequest) (*event.AddedResponse, error) {
	s.reporter.Increment(1)
	err := s.es.Add(ctx, convert.UserContextToDomain(in.UserContext), convert.EventsToDomain(in.Data))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &event.AddedResponse{}, nil
}

func (s *EventProtocService) UpdateSettings(ctx context.Context, in *event.UpdateSettingsRequest) (*event.SettingsUpdatedResponse, error) {
	s.reporter.Increment(1)
	err := s.es.UpdateSettings(ctx, convert.UserContextToDomain(in.UserContext), convert.UserEventSettingsToDomain(in.Settings))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &event.SettingsUpdatedResponse{}, nil
}

func (s *EventProtocService) NotifyComplete(ctx context.Context, in *event.NotifyCompleteRequest) (*event.NotifyCompletedResponse, error) {
	s.reporter.Increment(1)
	err := s.es.NotifyComplete(ctx, convert.UserContextToDomain(in.UserContext), in.StartTime, int(in.GroupID))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &event.NotifyCompletedResponse{}, nil
}
