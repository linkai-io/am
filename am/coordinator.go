package am

import (
	"context"
)

type CoordinatorStats struct {
	NumWorkers          int64
	NumActiveScanGroups int64
}

type ScanGroupStats struct {
}

type CoordinatorService interface {
	// externally accessable rpcs
	WorkerRegistration(ctx context.Context, workerID string) (*WorkerConfig, error)
	SystemStats(ctx context.Context, userContext UserContext) (*CoordinatorStats, error)
	GroupStats(ctx context.Context, userContext UserContext, scanGroupID int) (*ScanGroupStats, error)
	StartGroup(ctx context.Context, userContext UserContext, scanGroupID int) error
	StopGroup(ctx context.Context, userContext UserContext, scanGroupID int) error
	DeleteGroup(ctx context.Context, userContext UserContext, scanGroupID int) error

	// internal methods
	StartWorker() (string, error)
	StopWorker(workerID string) error
}
