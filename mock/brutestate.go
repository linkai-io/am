package mock

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

type BruteState struct {
	InitFn        func(config []byte) error
	InitFnInvoked bool

	DoBruteETLDFn      func(ctx context.Context, orgID, scanGroupID, expireSeconds int, maxAllowed int, etld string) (int, bool, error)
	DoBruteETLDInvoked bool

	DoBruteDomainFn      func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error)
	DoBruteDomainInvoked bool

	DoMutateDomainFn      func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error)
	DoMutateDomainInvoked bool

	SubscribeFn      func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error
	SubscribeInvoked bool

	GetGroupFn      func(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
	GetGroupInvoked bool
}

func (s *BruteState) Init(config []byte) error {
	return nil
}

func (s *BruteState) DoBruteETLD(ctx context.Context, orgID, scanGroupID, expireSeconds int, maxAllowed int, etld string) (int, bool, error) {
	s.DoBruteETLDInvoked = true
	return s.DoBruteETLDFn(ctx, orgID, scanGroupID, expireSeconds, maxAllowed, etld)
}

func (s *BruteState) DoBruteDomain(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	s.DoBruteDomainInvoked = true
	return s.DoBruteDomainFn(ctx, orgID, scanGroupID, expireSeconds, zone)
}

func (s *BruteState) DoMutateDomain(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	s.DoMutateDomainInvoked = true
	return s.DoMutateDomainFn(ctx, orgID, scanGroupID, expireSeconds, zone)
}

func (s *BruteState) Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
	s.SubscribeInvoked = true
	return s.SubscribeFn(ctx, onStartFn, onMessageFn, channels...)
}

func (s *BruteState) GetGroup(ctx context.Context, orgID int, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
	s.GetGroupInvoked = true
	return s.GetGroupFn(ctx, orgID, scanGroupID, wantModules)
}
