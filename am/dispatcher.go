package am

import "context"

const (
	DispatcherServiceKey = "dispatcherservice"
)

type DispatcherService interface {
	Init(config []byte) error
	PushAddresses(ctx context.Context, userContext UserContext, scanGroupID int) error
}
