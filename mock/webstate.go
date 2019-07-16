package mock

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

type WebState struct {
	InitFn        func(config []byte) error
	InitFnInvoked bool

	DoWebDomainFn      func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error)
	DoWebDomainInvoked bool

	SubscribeFn      func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error
	SubscribeInvoked bool

	GetGroupFn      func(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
	GetGroupInvoked bool

	GetPortResultsFn      func(ctx context.Context, orgID, scanGroupID int, host string) (*am.PortResults, error)
	GetPortResultsInvoked bool
}

func (s *WebState) Init(config []byte) error {
	return nil
}

func (s *WebState) DoWebDomain(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	s.DoWebDomainInvoked = true
	return s.DoWebDomainFn(ctx, orgID, scanGroupID, expireSeconds, zone)
}

func (s *WebState) Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
	s.SubscribeInvoked = true
	return s.SubscribeFn(ctx, onStartFn, onMessageFn, channels...)
}

func (s *WebState) GetGroup(ctx context.Context, orgID int, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
	s.GetGroupInvoked = true
	return s.GetGroupFn(ctx, orgID, scanGroupID, wantModules)
}

func (s *WebState) GetPortResults(ctx context.Context, orgID, scanGroupID int, host string) (*am.PortResults, error) {
	s.GetPortResultsInvoked = true
	return s.GetPortResultsFn(ctx, orgID, scanGroupID, host)
}
