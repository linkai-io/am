package cache

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/redisclient"
)

// CacheStater is for interfacing with a state management system (see coordinator/state/redis for implementation)
type CacheStater interface {
	GetAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[int64]*am.ScanGroupAddress, error)
	// Subscribe for updates
	// TODO: I know this is bad, an interface is reliant on an implementation (redisclient) change whenever.
	Subscribe(ctx context.Context, onStartFn redisclient.SubOnStart, onMessageFn redisclient.SubOnMessage, channels ...string) error
	// GetGroup returns the requested group with/without modules
	GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
}
