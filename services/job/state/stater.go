package state

import (
	"context"

	"gopkg.linkai.io/v1/repos/am/am"
)

// Stater is for interfacing with a state management system
// It is responsible for managing the life cycle of a job
type Stater interface {
	Init(config []byte) error
	Finalize(job *am.Job) error
	Start(ctx context.Context, userContext am.UserContext, jobID int64, group *am.ScanGroup) error
	Pause(ctx context.Context, userContext am.UserContext, jobID int64) error
	Resume(ctx context.Context, userContext am.UserContext, jobID int64) error
}
