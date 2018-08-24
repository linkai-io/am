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
	GroupStatus(ctx context.Context, userContext am.UserContext, scanGroupID int) (bool, am.GroupStatus, int64, error)
	Put(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) error
	Start(ctx context.Context, userContext am.UserContext, scanGroupID int) error
	Stop(ctx context.Context, userContext am.UserContext, scanGroupID int) error
	Delete(ctx context.Context, userContext am.UserContext, scanGroupID int) error
	PushAddresses(ctx context.Context, userContext am.UserContext, addresses []*am.ScanGroupAddress) error
}
