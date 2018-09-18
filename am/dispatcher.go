package am

import "context"

const (
	DispatcherServiceKey = "dispatcherservice"
)

type DispatcherService interface {
	PushAddresses(ctx context.Context, userContext UserContext, scanGroupID int) error
}
