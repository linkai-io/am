package coordinator

import (
	"context"
	"log"

	"github.com/pkg/errors"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/dispatcher"
	"github.com/linkai-io/am/services/coordinator/state"
)

var (
	modules            = []string{"ns", "dnsbrute", "port", "web"}
	ErrScanGroupPaused = errors.New("scan group is currently paused")
)

// Service for coordinating all scans
type Service struct {
	state             state.Stater
	dispatcherClients map[string]am.DispatcherService
	scanGroupClient   am.ScanGroupService
}

// New returns
func New(state state.Stater, scanGroupClient am.ScanGroupService) *Service {
	return &Service{
		state:             state,
		scanGroupClient:   scanGroupClient,
		dispatcherClients: make(map[string]am.DispatcherService),
	}
}

// Init by
func (s *Service) Init(config []byte) error {
	return nil
}

// StartGroup initializes state system if they do not exist, or updates with scan group details
func (s *Service) StartGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	oid, group, err := s.scanGroupClient.Get(ctx, userContext, scanGroupID)
	if err != nil {
		return err
	}

	if oid != userContext.GetOrgID() {
		return am.ErrOrgIDMismatch
	}

	if group.Paused {
		return ErrScanGroupPaused
	}

	exists, _, err := s.state.GroupStatus(ctx, userContext, scanGroupID)
	if err != nil {
		return errors.Wrap(err, "failed to get group status")
	}

	if !exists {
		// update/create configuration
		if err := s.state.Put(ctx, userContext, group); err != nil {
			return errors.Wrap(err, "failed to put new group")
		}
	}

	wantModules := true
	cachedGroup, err := s.state.GetGroup(ctx, userContext.GetOrgID(), scanGroupID, wantModules)
	if err != nil {
		return errors.Wrap(err, "failed to get cached group")
	}

	if cachedGroup.ModifiedTime < group.ModifiedTime {
		if err := s.state.Delete(ctx, userContext, group); err != nil {
			return errors.Wrap(err, "failed to delete group")
		}

		// update/create configuration
		if err := s.state.Put(ctx, userContext, group); err != nil {
			return errors.Wrap(err, "failed to update/put group")
		}
	}

	if err := s.state.Start(ctx, userContext, group.GroupID); err != nil {
		return errors.Wrap(err, "failed to start group")
	}

	return err
}

// Register the dispatcher and set status to registered in our state.
func (s *Service) Register(ctx context.Context, dispatcherAddress, dispatcherID string) error {
	log.Printf("dispatcher [%s] %s is now registered\n", dispatcherAddress, dispatcherID)
	dispatcherClient := dispatcher.New()
	if err := dispatcherClient.Init([]byte(dispatcherAddress)); err != nil {
		return err
	}
	s.dispatcherClients[dispatcherID] = dispatcherClient
	return nil
}
