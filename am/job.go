package am

import "context"

const (
	RNJobService = "lrn:service:jobservice:feature:service"
)

// Job represents a unit of work
type Job struct {
	ID        int64
	Config    *JobConfig
	StartTime int64
	EndTime   int64
}

type JobStatus struct {
	ID          int64
	Status      int32
	ModuleStats *ModuleStats
	Running     bool
}

// JobService interfaces with data store to manage jobs
type JobService interface {
	Jobs(ctx context.Context, orgID int64) ([]*Job, error)
	Add(ctx context.Context, orgID, userID int64, scanGroupID int64, name string) (int64, error)
	Get(ctx context.Context, orgID int64, jobID int64) *Job
	GetByName(ctx context.Context, orgID int64, name string) (*Job, error)
	Status(ctx context.Context, orgID int64, jobID int64) (*JobStatus, error)
	Stop(ctx context.Context, orgID int64, jobID int64) error
	Start(ctx context.Context, orgID int64, jobID int64) error
	UpdateStatus(ctx context.Context, orgID int64, jobID int64, jobStatus *JobStatus) error
}
