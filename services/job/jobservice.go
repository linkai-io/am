package job

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/auth"
	"gopkg.linkai.io/v1/repos/am/services/job/state"
)

var (
	ErrScanGroupPaused  = errors.New("scan group is currently paused")
	ErrUnknownJobStatus = errors.New("unknown job status returned")
	ErrJobInProgress    = errors.New("job for this scan group is already in progress")
)

// Service for interfacing with postgresql/rds
type Service struct {
	state           state.Stater
	pool            *pgx.ConnPool
	config          *pgx.ConnPoolConfig
	authorizer      auth.Authorizer
	addressClient   am.AddressService
	scanGroupClient am.ScanGroupService
}

// New returns an empty Service
func New(state state.Stater, addressClient am.AddressService, scanGroupClient am.ScanGroupService, authorizer auth.Authorizer) *Service {
	return &Service{state: state, authorizer: authorizer}
}

// Init by parsing the config and initializing the database pool
func (s *Service) Init(config []byte) error {
	var err error

	s.config, err = s.parseConfig(config)
	if err != nil {
		return err
	}

	if s.pool, err = pgx.NewConnPool(*s.config); err != nil {
		return err
	}

	return nil
}

// parseConfig parses the configuration options and validates they are sane.
func (s *Service) parseConfig(config []byte) (*pgx.ConnPoolConfig, error) {
	dbstring := string(config)
	if dbstring == "" {
		return nil, am.ErrEmptyDBConfig
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		return nil, am.ErrInvalidDBString
	}

	return &pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 50,
		AfterConnect:   s.afterConnect,
	}, nil
}

// afterConnect will iterate over prepared statements with keywords
func (s *Service) afterConnect(conn *pgx.Conn) error {
	for k, v := range queryMap {
		if _, err := conn.Prepare(k, v); err != nil {
			return err
		}
	}
	return nil
}

// IsAuthorized checks if an action is allowed by a particular user
func (s *Service) IsAuthorized(ctx context.Context, userContext am.UserContext, resource, action string) bool {
	if err := s.authorizer.IsUserAllowed(userContext.GetOrgID(), userContext.GetUserID(), resource, action); err != nil {
		return false
	}
	return true
}

// Start a job for a scan group
// 1. First check if this scan group is already running a job and this scan group is not paused.
// 2. Create the job in the jobs table
// 3. Add necessary events
// 4. Pull out scan group config and push to redis/stater
func (s *Service) Start(ctx context.Context, userContext am.UserContext, groupID int) (oid int, jobID int64, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobServiceLifeCycle, "create") {
		return 0, 0, am.ErrUserNotAuthorized
	}

	var tx *pgx.Tx

	group, err := s.getScanGroup(ctx, userContext, groupID)
	if err != nil {
		return 0, 0, err
	}

	tx, err = s.pool.BeginEx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback() // safe to call as no-op on success

	if err := s.checkLastJob(userContext, groupID, tx); err != nil {
		return 0, 0, err
	}

	now := time.Now().UnixNano()

	err = tx.QueryRow("startJob", userContext.GetOrgID(), groupID, now, int(am.JobStarted)).Scan(&oid, &jobID)
	if err != nil {
		return 0, 0, err
	}

	_, err = tx.Exec("createJobEvent", userContext.GetOrgID(), jobID, userContext.GetUserID(), now, "job started", "job service")
	if err != nil {
		return 0, 0, err
	}

	if err := s.state.Start(ctx, userContext, jobID, group); err != nil {
		return 0, 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, 0, err
	}

	return oid, jobID, nil
}

// checkLastJob ensures there isn't any running jobs or the last job was paused.
func (s *Service) checkLastJob(userContext am.UserContext, groupID int, tx *pgx.Tx) error {
	job := &am.Job{}

	err := tx.QueryRow("getLastJob", userContext.GetOrgID(), groupID).Scan(&job.JobID, &job.OrgID, &job.GroupID, &job.JobTimestamp, &job.JobStatus)
	if err != nil {
		return err
	}

	if am.JobStatus(job.JobStatus) == am.JobStarted || am.JobStatus(job.JobStatus) == am.JobPaused {
		return ErrJobInProgress
	}
	return nil
}

// getScanGroup from the scanGroupClient and ensure the entire scan group is not in a paused state.
func (s *Service) getScanGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) (*am.ScanGroup, error) {
	oid, group, err := s.scanGroupClient.Get(ctx, userContext, scanGroupID)
	if err != nil {
		return nil, err
	}

	if oid != userContext.GetOrgID() || group.OrgID != userContext.GetOrgID() {
		return nil, am.ErrOrgIDMismatch
	}

	if group.Paused {
		return nil, ErrScanGroupPaused
	}

	return group, nil
}

// Pause the current running job via JobID.
func (s *Service) Pause(ctx context.Context, userContext am.UserContext, jobID int64) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobServiceLifeCycle, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	return 0, nil
}

// Resume the current running job (if paused).
func (s *Service) Resume(ctx context.Context, userContext am.UserContext, jobID int64) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobServiceLifeCycle, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	return 0, nil
}

// PauseGroup pauses the entire scan group from running as well as any current running scans for that group.
func (s *Service) PauseGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobServiceLifeCycle, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	return 0, nil
}

// ResumeGroup resumes the entire scan group as well as any currently paused scans that were running.
func (s *Service) ResumeGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobServiceLifeCycle, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	return 0, nil
}

// Cancel the current running job.
func (s *Service) Cancel(ctx context.Context, userContext am.UserContext, jobID int64) (oid int, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobServiceLifeCycle, "update") {
		return 0, am.ErrUserNotAuthorized
	}
	return 0, nil
}

// Get the job details via the supplied jobID.
func (s *Service) Get(ctx context.Context, userContext am.UserContext, jobID int64) (oid int, job *am.Job, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobService, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return 0, nil, nil
}

// List all jobs that match the specified filter.
func (s *Service) List(ctx context.Context, userContext am.UserContext, filter *am.JobFilter) (oid int, jobs []*am.Job, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobService, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return 0, nil, nil
}

// CreateEvent for a job.
func (s *Service) CreateEvent(ctx context.Context, userContext am.UserContext, jobEvent *am.JobEvent) (oid int, eventID int64, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobServiceEvents, "create") {
		return 0, 0, am.ErrUserNotAuthorized
	}

	err = s.pool.QueryRow("createJobEvent", jobEvent.OrgID, jobEvent.JobID, jobEvent.EventUserID, jobEvent.EventTime, jobEvent.EventDescription, jobEvent.EventFrom).Scan(&oid, &eventID)

	return oid, eventID, err
}

// GetEvents that match the specified filter.
func (s *Service) GetEvents(ctx context.Context, userContext am.UserContext, filter *am.JobEventFilter) (oid int, jobEvents []*am.JobEvent, err error) {
	if !s.IsAuthorized(ctx, userContext, am.RNJobServiceEvents, "read") {
		return 0, nil, am.ErrUserNotAuthorized
	}
	return 0, nil, nil
}
