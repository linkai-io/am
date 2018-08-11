package am

import "context"

const (
	RNJobService          = "lrn:service:jobservice:feature:service"
	RNJobServiceLifeCycle = "lrn:service:jobservice:feature:lifecycle"
	RNJobServiceEvents    = "lrn:service:jobservice:feature:events"
)

type JobStatus int

var (
	JobStarted  JobStatus = 1
	JobPaused   JobStatus = 2
	JobCanceled JobStatus = 3
	JobFinished JobStatus = 4
)

var JobStatusMap = map[JobStatus]string{
	1: "started",
	2: "paused",
	3: "canceled",
	4: "finished",
}

// Job represents a unit of work
type Job struct {
	OrgID        int   `json:"org_id"`
	JobID        int64 `json:"job_id"`
	GroupID      int   `json:"group_id"`
	JobTimestamp int64 `json:"job_timestamp"`
	JobStatus    int   `json:"job_status"`
}

type JobEvent struct {
	EventID          int64  `json:"event_id"`
	OrgID            int    `json:"org_id"`
	JobID            int64  `json:"job_id"`
	EventUserID      int    `json:"event_user_id"`
	EventTime        int64  `json:"event_time"`
	EventDescription string `json:"event_description"`
	EventFrom        string `json:"event_from"`
}

type JobFilter struct {
	Start    int   `json:"start"`
	Limit    int   `json:"limit"`
	ByJobID  bool  `json:"by_job_id"`
	JobID    int64 `json:"job_id,omitempty"`
	ByStatus bool  `json:"by_status"`
	StatusID int   `json:"status_id"`
}

type JobEventFilter struct {
	Start int `json:"start"`
	Limit int `json:"limit"`
}

// JobService interfaces with data store to manage jobs
type JobService interface {
	Start(ctx context.Context, userContext UserContext, scanGroupID int) (oid int, jobID int64, err error)
	Pause(ctx context.Context, userContext UserContext, jobID int64) (oid int, err error)
	Resume(ctx context.Context, userContext UserContext, jobID int64) (oid int, err error)
	Cancel(ctx context.Context, userContext UserContext, jobID int64) (oid int, err error)
	Get(ctx context.Context, userContext UserContext, jobID int64) (oid int, job *Job, err error)                               // config
	List(ctx context.Context, userContext UserContext, filter *JobFilter) (oid int, jobs []*Job, err error)                     // list jobs via filter
	CreateEvent(ctx context.Context, userContext UserContext, jobID int64) (oid int, eventID int64, err error)                  //
	GetEvents(ctx context.Context, userContext UserContext, filter *JobEventFilter) (oid int, jobEvents []*JobEvent, err error) //
}
