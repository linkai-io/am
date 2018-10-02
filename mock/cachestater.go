package mock

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

type CacheStater struct {
	GetAddressesFn      func(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[int64]*am.ScanGroupAddress, error)
	GetAddressesInvoked bool

	SubscribeFn      func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error
	SubscribeInvoked bool

	GetGroupFn      func(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
	GetGroupInvoked bool
}

func (c *CacheStater) GetAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[int64]*am.ScanGroupAddress, error) {
	c.GetAddressesInvoked = true
	return c.GetAddressesFn(ctx, userContext, scanGroupID, limit)
}

func (c *CacheStater) Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
	c.SubscribeInvoked = true
	return c.SubscribeFn(ctx, onStartFn, onMessageFn, channels...)
}

func (c *CacheStater) GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
	c.GetGroupInvoked = true
	return c.GetGroupFn(ctx, orgID, scanGroupID, wantModules)
}
