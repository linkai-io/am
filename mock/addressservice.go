package mock

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
)

type AddressService struct {
	InitFn        func(config []byte) error
	InitFnInvoked bool

	GetFn      func(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, addresses []*am.ScanGroupAddress, err error)
	GetInvoked bool

	GetHostListFn      func(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, hosts []*am.ScanGroupHostList, err error)
	GetHostListInvoked bool

	CountFn      func(ctx context.Context, userContext am.UserContext, groupID int) (oid int, count int, err error)
	CountInvoked bool

	UpdateFn      func(ctx context.Context, userContext am.UserContext, addresses map[string]*am.ScanGroupAddress) (oid int, count int, err error)
	UpdateInvoked bool

	DeleteFn      func(ctx context.Context, userContext am.UserContext, groupID int, addressIDs []int64) (oid int, err error)
	DeleteInvoked bool

	IgnoreFn      func(ctx context.Context, userContext am.UserContext, groupID int, addressIDs []int64, ignoreValue bool) (oid int, err error)
	IgnoreInvoked bool

	OrgStatsFn      func(ctx context.Context, userContext am.UserContext) (int, []*am.ScanGroupAddressStats, error)
	OrgStatsInvoked bool

	GroupStatsFn      func(ctx context.Context, userContext am.UserContext, groupID int) (int, *am.ScanGroupAddressStats, error)
	GroupStatsInvoked bool

	ArchiveFn      func(ctx context.Context, userContext am.UserContext, group *am.ScanGroup, archiveTime time.Time) (int, int, error)
	ArchiveInvoked bool

	UpdateHostPortsFn      func(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, portResults *am.PortResults) (oid int, err error)
	UpdateHostPortsInvoked bool

	GetPortsFn      func(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, portResults []*am.PortResults, err error)
	GetPortsInvoked bool
}

func (c *AddressService) Init(config []byte) error {
	return nil
}

func (c *AddressService) Get(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, addresses []*am.ScanGroupAddress, err error) {
	c.GetInvoked = true
	return c.GetFn(ctx, userContext, filter)
}

func (c *AddressService) GetHostList(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, hosts []*am.ScanGroupHostList, err error) {
	c.GetHostListInvoked = true
	return c.GetHostListFn(ctx, userContext, filter)
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

func (c *AddressService) Ignore(ctx context.Context, userContext am.UserContext, groupID int, addressIDs []int64, ignoreValue bool) (oid int, err error) {
	c.IgnoreInvoked = true
	return c.IgnoreFn(ctx, userContext, groupID, addressIDs, ignoreValue)
}

func (c *AddressService) OrgStats(ctx context.Context, userContext am.UserContext) (int, []*am.ScanGroupAddressStats, error) {
	c.OrgStatsInvoked = true
	return c.OrgStatsFn(ctx, userContext)
}

func (c *AddressService) GroupStats(ctx context.Context, userContext am.UserContext, groupID int) (int, *am.ScanGroupAddressStats, error) {
	c.GroupStatsInvoked = true
	return c.GroupStatsFn(ctx, userContext, groupID)
}

func (c *AddressService) Archive(ctx context.Context, userContext am.UserContext, group *am.ScanGroup, archiveTime time.Time) (int, int, error) {
	c.ArchiveInvoked = true
	return c.ArchiveFn(ctx, userContext, group, archiveTime)
}

func (c *AddressService) UpdateHostPorts(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, portResults *am.PortResults) (oid int, err error) {
	c.UpdateHostPortsInvoked = true
	return c.UpdateHostPortsFn(ctx, userContext, address, portResults)
}

func (c *AddressService) GetPorts(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, portResults []*am.PortResults, err error) {
	c.GetPortsInvoked = true
	return c.GetPortsFn(ctx, userContext, filter)
}
