package mock

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

type NSState struct {
	InitFn        func(config []byte) error
	InitFnInvoked bool

	DoNSRecordsFn      func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error)
	DoNSRecordsInvoked bool

	GetAddressesFn      func(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[int64]*am.ScanGroupAddress, error)
	GetAddressesInvoked bool

	SubscribeFn      func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error
	SubscribeInvoked bool

	GetGroupFn      func(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
	GetGroupInvoked bool
}

func (s *NSState) Init(config []byte) error {
	return nil
}

func (s *NSState) DoNSRecords(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	s.DoNSRecordsInvoked = true
	return s.DoNSRecordsFn(ctx, orgID, scanGroupID, expireSeconds, zone)
}

func (s *NSState) Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
	s.SubscribeInvoked = true
	return s.SubscribeFn(ctx, onStartFn, onMessageFn, channels...)
}

func (s *NSState) GetGroup(ctx context.Context, orgID int, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
	s.GetGroupInvoked = true
	return s.GetGroupFn(ctx, orgID, scanGroupID, wantModules)
}

func (s *NSState) GetAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[int64]*am.ScanGroupAddress, error) {
	s.GetAddressesInvoked = true
	return s.GetAddressesFn(ctx, userContext, scanGroupID, limit)
}
