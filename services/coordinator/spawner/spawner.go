package spawner

import (
	"context"

	"github.com/linkai-io/am/am"
)

type WorkerData struct {
	ID                string
	Address           string
	ModuleType        am.ModuleType
	Registered        bool
	StartTime         int64
	MessagesProcessed int64
	RequeueCount      int64
}

type Spawner interface {
	Spawn(ctx context.Context, moduleType am.ModuleType) (*WorkerData, error)
	Kill(ctx context.Context, workerID string) error
}
