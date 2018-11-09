package state

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/state"
)

// Stater is for interfacing with a state management system (see pkg/state/redis/redis.go for implementation)
type Stater interface {
	// TODO: Add WebDomains logic so we can search for domains that don't match the etld we are analyzing, but are in
	// our scan group and 'verified' as owned.

	// WebDomains returns verified domains that are apart of this scan group (added from AddWebDomain)
	//GetWebDomains(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) ([]string, error)

	// AddWebDomain adds this zone/tld to our list of web domains that are safe to search for
	//AddWebDomain(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error)

	// DoWebDomain until intercepting requests and injecting IPs is fixed, only bother doing web analysis one per domain
	DoWebDomain(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error)

	// Subscribe for updates
	Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error

	// GetGroup returns the requested group with/without modules
	GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error)
}
