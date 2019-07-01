package state

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

// Stater is for interfacing with a state management system
// It is responsible for managing the life cycle of scangroups
// and tracking global scan state
type Stater interface {
	// Subscribe for updates
	Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error
	GroupStatus(ctx context.Context, userContext am.UserContext, scanGroupID int) (bool, am.GroupStatus, error)
	GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
	Stop(ctx context.Context, userContext am.UserContext, scanGroupID int) error
	PutAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses []*am.ScanGroupAddress) error
	PutAddressMap(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses map[string]*am.ScanGroupAddress) error
	PopAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[string]*am.ScanGroupAddress, error)
	FilterNew(ctx context.Context, orgID, scanGroupID int, addresses map[string]*am.ScanGroupAddress) (map[string]*am.ScanGroupAddress, error)
	// DoPortScan determines if we should port scan this host (or ip)
	DoPortScan(ctx context.Context, orgID, scanGroupID int, expireSeconds int, host string) (bool, error)
}
