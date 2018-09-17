package coordinator

import (
	"context"

	"github.com/linkai-io/am/am"

	"github.com/linkai-io/am/services/coordinator/spawner"
	"github.com/linkai-io/am/services/coordinator/state"
)

// DispatcherCoordinator for coordinating the lifecycle of workers
type DispatcherCoordinator struct {
	env     string
	region  string
	state   state.Stater
	spawner *spawner.Spawn
}

// NewDispatcherCoordinator for coordinating the work of workers
func NewDispatcherCoordinator(env, region string, stater state.Stater) *DispatcherCoordinator {
	dc := &DispatcherCoordinator{state: stater, env: env, region: region}
	dc.spawner = spawner.New(env, region)
	return dc
}

// Spawn count new worker(s) via spawner of module type for queue.
func (dc *DispatcherCoordinator) Spawn(ctx context.Context, moduleType am.ModuleType, count int) error {
	return dc.Spawn(ctx, moduleType, count)
}

// Kill the worker by the provided dispatcherID
func (dc *DispatcherCoordinator) Kill(ctx context.Context, dispatcherID string) error {
	return nil
}

// Register the worker and set status to registered in our state.
func (dc *DispatcherCoordinator) Register(ctx context.Context, dispatcherID string) error {
	return nil
}
