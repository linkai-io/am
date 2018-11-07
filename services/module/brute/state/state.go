package state

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

// Stater is for interfacing with a state management system (see pkg/state/redis/redis.go for implementation)
type Stater interface {
	// Initialize the state system
	Init(config []byte) error
	// DoBruteETLD increments a counter so we don't overload an etld
	DoBruteETLD(ctx context.Context, orgID, scanGroupID, expireSeconds int, maxAllowed int, etld string) (int, bool, error)
	// DoBruteDomain returns true if we should brute force the zone and sets a key in redis. Otherwise
	// returns false stating we've already brute forced it (until expireSeconds)
	DoBruteDomain(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error)
	// DoMutateDomain returns true if we should mutate the zone and sets a key in redis. Otherwise
	// returns false stating we've already brute forced it (until expireSeconds)
	DoMutateDomain(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error)
	// Subscribe for updates
	Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error
	// GetGroup returns the requested group with/without modules
	GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
}
