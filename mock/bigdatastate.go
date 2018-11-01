package mock

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

type BigDataState struct {
	InitFn        func(config []byte) error
	InitFnInvoked bool

	DoCTDomainFn      func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error)
	DoCTDomainInvoked bool

	SubscribeFn      func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error
	SubscribeInvoked bool

	GetGroupFn      func(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
	GetGroupInvoked bool
}

func (s *BigDataState) Init(config []byte) error {
	return nil
}

func (s *BigDataState) DoCTDomain(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	s.DoCTDomainInvoked = true
	return s.DoCTDomainFn(ctx, orgID, scanGroupID, expireSeconds, zone)
}

func (s *BigDataState) Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
	s.SubscribeInvoked = true
	return s.SubscribeFn(ctx, onStartFn, onMessageFn, channels...)
}

func (s *BigDataState) GetGroup(ctx context.Context, orgID int, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
	s.GetGroupInvoked = true
	return s.GetGroupFn(ctx, orgID, scanGroupID, wantModules)
}
