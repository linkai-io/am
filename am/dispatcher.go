package am

import (
	"context"
)

const (
	DispatcherServiceKey = "dispatcherservice"
)

// DispatcherService handles dispatching scan group addresses to the analysis modules
type DispatcherService interface {
	Init(config []byte) error
	PushAddresses(ctx context.Context, userContext UserContext, scanGroupID int) error
}
