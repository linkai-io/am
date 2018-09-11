package state

import (
	"context"

	"github.com/linkai-io/am/pkg/redisclient"
)

// Stater is for interfacing with a state management system
type Stater interface {
	// Initialize the state system needs org_id and supporting connection details
	Init(config []byte) error
	// safely check if we should lookup records for this zone
	DoNSRecords(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error)
	// Checks if the zone is OK to be been analyzed.
	IsValid(zone string) bool
	// Subscribe for updates
	// TODO: I know this is bad, an interface is reliant on an implementation (redisclient) change whenever.
	Subscribe(ctx context.Context, onStartFn redisclient.SubOnStart, onMessageFn redisclient.SubOnMessage, channels ...string) error
}
