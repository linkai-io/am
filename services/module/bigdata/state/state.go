package state

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

// Stater is for interfacing with a state management system (see pkg/state/redis/redis.go for implementation)
type Stater interface {

	// DoCTDomain checks if we should check our database and bigquery for new certificate transparency results
	DoCTDomain(ctx context.Context, orgID, scanGroupID int, expireSeconds int, etld string) (bool, error)

	// Subscribe for updates
	Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error

	// GetGroup returns the requested group with/without modules
	GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
}
