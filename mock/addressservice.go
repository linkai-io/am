package mock

import (
	"context"

	"github.com/linkai-io/am/am"
)

type AddressService struct {
	GetFn      func(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, addresses []*am.ScanGroupAddress, err error)
	GetInvoked bool

	CountFn      func(ctx context.Context, userContext am.UserContext, groupID int) (oid int, count int, err error)
	CountInvoked bool

	UpdateFn      func(ctx context.Context, userContext am.UserContext, addresses map[string]*am.ScanGroupAddress) (oid int, count int, err error)
	UpdateInvoked bool

	DeleteFn      func(ctx context.Context, userContext am.UserContext, groupID int, addressIDs []int64) (oid int, err error)
	DeleteInvoked bool
}

func (c *AddressService) Get(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, addresses []*am.ScanGroupAddress, err error) {
	c.GetInvoked = true
	return c.GetFn(ctx, userContext, filter)
}

func (c *AddressService) Count(ctx context.Context, userContext am.UserContext, groupID int) (oid int, count int, err error) {
	c.CountInvoked = true
	return c.CountFn(ctx, userContext, groupID)
}

func (c *AddressService) Update(ctx context.Context, userContext am.UserContext, addresses map[string]*am.ScanGroupAddress) (oid int, count int, err error) {
	c.UpdateInvoked = true
	return c.UpdateFn(ctx, userContext, addresses)
}

func (c *AddressService) Delete(ctx context.Context, userContext am.UserContext, groupID int, addressIDs []int64) (oid int, err error) {
	c.DeleteInvoked = true
	return c.DeleteFn(ctx, userContext, groupID, addressIDs)
}
