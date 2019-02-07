package am

import (
	"context"
)

const (
	CoordinatorServiceKey = "coordinatorservice"
)

type ScanGroupStats struct {
}

type CoordinatorService interface {
	Init(config []byte) error
	// externally accessable rpcs
	//GroupStats(ctx context.Context, userContext UserContext, scanGroupID int) (*ScanGroupStats, error)
	StartGroup(ctx context.Context, userContext UserContext, scanGroupID int) error
	StopGroup(ctx context.Context, userContext UserContext, scanGroupID int) (string, error)
	//DeleteGroup(ctx context.Context, userContext UserContext, scanGroupID int) error

	// internal methods
	//StartWorker() (string, error)
	//StopWorker(workerID string) error
}
