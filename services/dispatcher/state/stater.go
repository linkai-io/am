package state

import (
	"context"

	"github.com/linkai-io/am/am"
)

// Stater is for interfacing with a state management system
// It is responsible for managing the life cycle of scangroups
// and tracking global scan state
type Stater interface {
	Init(config []byte) error
	GroupStatus(ctx context.Context, userContext am.UserContext, scanGroupID int) (bool, am.GroupStatus, error)
	GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
	Stop(ctx context.Context, userContext am.UserContext, scanGroupID int) error
	PutAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses []*am.ScanGroupAddress) error
	PutAddressMap(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses map[string]*am.ScanGroupAddress) error
	PopAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[string]*am.ScanGroupAddress, error)
	FilterNew(ctx context.Context, orgID, scanGroupID int, addresses map[string]*am.ScanGroupAddress) (map[string]*am.ScanGroupAddress, error)
}
