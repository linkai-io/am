package spawner

import (
	"context"

	"github.com/linkai-io/am/am"
)

type AWSSpawner struct {
	env    string
	region string
}

func NewAWSSpawner(env, region string) *AWSSpawner {
	return &AWSSpawner{env: env, region: region}
}

func (s *AWSSpawner) Spawn(ctx context.Context, moduleType am.ModuleType) (*WorkerData, error) {
	return nil, nil
}

func (s *AWSSpawner) Kill(ctx context.Context, workerID string) error {
	return nil
}
