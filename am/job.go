package am

// Job represents a unit of work
type Job struct {
	Config    *JobConfig
	ID        []byte
	StartTime int64
	EndTime   int64
}

type JobStatus struct {
	Status      int
	ModuleStats *ModuleStats
	Running     bool
}

// JobService interfaces with data store to manage jobs
type JobService interface {
	Jobs(orgID int64) ([]*Job, error)
	Add(orgID int64, userID int64, name string, inputID int64) ([]byte, error)
	Get(orgID int64, jobID []byte) *Job
	GetByName(orgID int64, name string) (*Job, error)
	Status(orgID int64, jobID []byte) (*JobStatus, error)
	Stop(orgID int64, jobID []byte) error
	Start(orgID int64, jobID []byte) error
	UpdateStatus(job *Job, jobStatus *JobStatus) error
}
