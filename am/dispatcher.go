package am

import "context"

type DispatcherService interface {
	PushAddresses(ctx context.Context, userContext UserContext, scanGroupID int) error
}
