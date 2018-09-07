package coordinator

import (
	"context"

	"github.com/linkai-io/am/am"

	"github.com/linkai-io/am/services/coordinator/spawner"
	"github.com/linkai-io/am/services/coordinator/state"
)

// WorkerCoordinator for coordinating the lifecycle of workers
type WorkerCoordinator struct {
	env     string
	region  string
	state   state.Stater
	spawner spawner.Spawner
}

// NewWorkerCoordinator for coordinating the work of workers
func NewWorkerCoordinator(env, region string, stater state.Stater) *WorkerCoordinator {
	wc := &WorkerCoordinator{state: stater, env: env, region: region}
	wc.spawner = spawner.New(env, region)
	return wc
}

// Spawn N new worker(s) via spawner of module type for queue.
func (wc *WorkerCoordinator) Spawn(ctx context.Context, moduleType am.ModuleType, queue string, count int) error {
}

// Kill the worker by the provided workerID
func (wc *WorkerCoordinator) Kill(ctx context.Context, workerID string) error {
	return nil
}

// KillAll workers for a scan group ID
func (wc *WorkerCoordinator) KillAll(ctx context.Context, scanGroupID int) {

}

// Register the worker and set status to registered in our state.
func (wc *WorkerCoordinator) Register(ctx context.Context, workerID string) error {
	return nil
}
