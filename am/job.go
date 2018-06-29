package am

import "context"

// Job represents a unit of work
type Job struct {
	ID        []byte
	Config    *JobConfig
	StartTime int64
	EndTime   int64
}

type JobStatus struct {
	ID          []byte
	Status      int32
	ModuleStats *ModuleStats
	Running     bool
}

// JobService interfaces with data store to manage jobs
type JobService interface {
	Jobs(ctx context.Context, orgID int64) ([]*Job, error)
	Add(ctx context.Context, orgID, userID int64, scanGroupID int64, name string) ([]byte, error)
	Get(ctx context.Context, orgID int64, jobID []byte) *Job
	GetByName(ctx context.Context, orgID int64, name string) (*Job, error)
	Status(ctx context.Context, orgID int64, jobID []byte) (*JobStatus, error)
	Stop(ctx context.Context, orgID int64, jobID []byte) error
	Start(ctx context.Context, orgID int64, jobID []byte) error
	UpdateStatus(ctx context.Context, orgID int64, jobID []byte, jobStatus *JobStatus) error
}
