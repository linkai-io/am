package module

import (
	"context"

	"github.com/linkai-io/am/am"
)

// Service for interfacing with coordinator/modules
type Service struct {
	config     *am.WorkerConfig
	statistics *am.WorkerReport
}

// New returns an empty Service
func New() *Service {
	s := &Service{}
	s.statistics = &am.WorkerReport{}
	return &Service{}
}

// Init by parsing the config and initializing
func (s *Service) Init(config []byte) error {
	var err error

	if err = s.parseConfig(config); err != nil {
		return err
	}

	return nil
}

// parseConfig parses the configuration options and validates they are sane.
func (s *Service) parseConfig(config []byte) error {
	return nil
}

// register this module servicer with the coordinator
func (s *Service) register(ctx context.Context) error {
	return nil
}

// Report on statistics of all running workers
func (s *Service) Report(ctx context.Context) (*am.WorkerReport, error) {
	return nil, nil
}

// Heartbeat if alive
func (s *Service) Heartbeat(ctx context.Context) bool {
	return true
}

func (s *Service) Shutdown(ctx context.Context) error {
	return nil
}
