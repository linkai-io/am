package job

import (
	"gopkg.linkai.io/v1/repos/am/services/job/state"
	"gopkg.linkai.io/v1/repos/am/services/job/store"
)

// Service is the job service which manages both the backend data store
// and the job state via a state manager
type Service struct {
	store store.Storer // store handles storing any job related data as well as lifecycle events pertinent to the front end
	state state.Stater // state handles job lifecycle events and data
}

// New creates a new Job Service.
func New(dataStore store.Storer, stateManager state.Stater) *Service {
	return &Service{store: dataStore, state: stateManager}
}

// Init the job service by initializing the data store and state managers.
func (s *Service) Init(config []byte) error {
	if err := s.state.Init(config); err != nil {
		return err
	}

	if err := s.store.Init(config); err != nil {
		return err
	}
	return nil
}
