package state

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

// Stater is for interfacing with a state management system (see coordinator/state/redis for implementation)
type Stater interface {
	// safely check if we should lookup records for this zone
	DoNSRecords(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error)
	// Subscribe for updates
	Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error
	// GetGroup returns the requested group with/without modules
	GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
}
