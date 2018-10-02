package cache

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

// CacheStater is for interfacing with a state management system (see coordinator/state/redis for implementation)
type CacheStater interface {
	// Subscribe for updates
	Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error
	// GetGroup returns the requested group with/without modules
	GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
}
