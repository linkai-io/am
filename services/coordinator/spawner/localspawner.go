package spawner

import (
	"context"

	"github.com/linkai-io/am/am"
)

type LocalSpawner struct {
	env    string
	region string
}

func NewLocalSpawner(env, region string) *LocalSpawner {
	return &LocalSpawner{env: env, region: region}
}

func (s *LocalSpawner) Spawn(ctx context.Context, moduleType am.ModuleType) (*WorkerData, error) {
	switch moduleType {
	case am.NSModule:

	}
	return nil, nil
}

func (s *LocalSpawner) Kill(ctx context.Context, workerID string) error {
	return nil
}
