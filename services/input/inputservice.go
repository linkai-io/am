package input

import (
	"github.com/linkai-io/am/am"
)

// Service for handling input for scan groups
type Service struct {
	scanGroupClient am.ScanGroupService
}

// New returns an empty Service
func New(scanGroupClient am.ScanGroupService) *Service {
	return &Service{scanGroupClient: scanGroupClient}
}

// Init ...
func (s *Service) Init(config []byte) error {
	return nil
}
