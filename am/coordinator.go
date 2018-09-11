package am

import (
	"context"
)

type ScanGroupStats struct {
}

type CoordinatorService interface {
	// externally accessable rpcs
	Register(ctx context.Context, dispatcherID string) error
	GroupStats(ctx context.Context, userContext UserContext, scanGroupID int) (*ScanGroupStats, error)
	StartGroup(ctx context.Context, userContext UserContext, scanGroupID int) error
	StopGroup(ctx context.Context, userContext UserContext, scanGroupID int) error
	DeleteGroup(ctx context.Context, userContext UserContext, scanGroupID int) error

	// internal methods
	StartWorker() (string, error)
	StopWorker(workerID string) error
}
