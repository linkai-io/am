package job

import (
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/services/job/state"
	"gopkg.linkai.io/v1/repos/am/services/job/store"
)

// Service is the job service which manages both the backend data store
// and the job state via a state manager
type Service struct {
	store store.Storer          // store handles any job related data as well as lifecycle events pertinent to the front end
	state state.Stater          // state handles job lifecycle events and data
	input am.InputReaderService // for accessing the org/group/job input list
}

// New creates a new Job Service.
func New(dataStore store.Storer, stateManager state.Stater) *Service {
	return &Service{store: dataStore, state: stateManager}
}

// Init the job service by initializing the dependent services.
func (s *Service) Init(config []byte) error {
	if err := s.state.Init(config); err != nil {
		return err
	}

	if err := s.store.Init(config); err != nil {
		return err
	}

	if err := s.input.Init(config); err != nil {
		return err
	}
	return nil
}

func (s *Service) Jobs(orgID int64) ([]*am.Job, error) {
	return nil, nil
}

func (s *Service) Add(orgID int64, userID int64, name string, inputID int64) ([]byte, error) {
	return nil, nil
}

func (s *Service) Get(orgID int64, jobID []byte) *am.Job {
	return nil
}

func (s *Service) GetByName(orgID int64, name string) (*am.Job, error) {
	return nil, nil
}

func (s *Service) Status(orgID int64, jobID []byte) (*am.JobStatus, error) {
	return nil, nil
}

func (s *Service) Stop(orgID int64, jobID []byte) error {
	return nil
}

func (s *Service) Start(orgID int64, jobID []byte) error {
	return nil
}

func (s *Service) UpdateStatus(job *am.Job, jobStatus *am.JobStatus) error {
	return nil
}
