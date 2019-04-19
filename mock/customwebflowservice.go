package mock

import (
	"context"

	"github.com/linkai-io/am/am"
)

type CustomWebFlowService struct {
	InitFn        func(config []byte) error
	InitFnInvoked bool

	CreateFn        func(ctx context.Context, userContext am.UserContext, config *am.CustomWebFlowConfig) (int, error)
	CreateFnInvoked bool

	UpdateFn        func(ctx context.Context, userContext am.UserContext, config *am.CustomWebFlowConfig) (int, error)
	UpdateFnInvoked bool

	DeleteFn        func(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error)
	DeleteFnInvoked bool

	StartFn        func(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error)
	StartFnInvoked bool

	StopFn        func(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error)
	StopFnInvoked bool

	GetStatusFn        func(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, *am.CustomWebStatus, error)
	GetStatusFnInvoked bool

	GetResultsFn        func(ctx context.Context, userContext am.UserContext, filter *am.CustomWebFilter) (int, []*am.CustomWebFlowResults, error)
	GetResultsFnInvoked bool
}

// Init by parsing the config and initializing the database pool
func (s *CustomWebFlowService) Init(config []byte) error {
	s.InitFnInvoked = true
	return s.InitFn(config)
}

func (s *CustomWebFlowService) Create(ctx context.Context, userContext am.UserContext, config *am.CustomWebFlowConfig) (int, error) {
	s.CreateFnInvoked = true
	return s.CreateFn(ctx, userContext, config)
}

func (s *CustomWebFlowService) Update(ctx context.Context, userContext am.UserContext, config *am.CustomWebFlowConfig) (int, error) {
	s.UpdateFnInvoked = true
	return s.UpdateFn(ctx, userContext, config)
}

func (s *CustomWebFlowService) Delete(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error) {
	s.DeleteFnInvoked = true
	return s.DeleteFn(ctx, userContext, webFlowID)
}

func (s *CustomWebFlowService) Start(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error) {
	s.StartFnInvoked = true
	return s.StartFn(ctx, userContext, webFlowID)
}

func (s *CustomWebFlowService) Stop(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, error) {
	s.StopFnInvoked = true
	return s.StopFn(ctx, userContext, webFlowID)
}

func (s *CustomWebFlowService) GetStatus(ctx context.Context, userContext am.UserContext, webFlowID int32) (int, *am.CustomWebStatus, error) {
	s.GetStatusFnInvoked = true
	return s.GetStatusFn(ctx, userContext, webFlowID)
}

func (s *CustomWebFlowService) GetResults(ctx context.Context, userContext am.UserContext, filter *am.CustomWebFilter) (int, []*am.CustomWebFlowResults, error) {
	s.GetResultsFnInvoked = true
	return s.GetResultsFn(ctx, userContext, filter)
}
