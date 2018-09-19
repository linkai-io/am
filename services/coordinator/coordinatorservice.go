package coordinator

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/linkai-io/am/pkg/retrier"

	"github.com/pkg/errors"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/dispatcher"
	"github.com/linkai-io/am/services/coordinator/state"
)

var (
	modules                  = []string{"ns", "dnsbrute", "port", "web"}
	ErrScanGroupPaused       = errors.New("scan group is currently paused")
	ErrNoDispatcherConnected = errors.New("service unavailable, no dispatchers connected")
)

// Service for coordinating all scans
type Service struct {
	loadBalancerAddr string
	state            state.Stater
	dispatcherClient am.DispatcherService
	scanGroupClient  am.ScanGroupService
	connected        int32
}

// New returns
func New(state state.Stater, scanGroupClient am.ScanGroupService) *Service {
	return &Service{
		state:           state,
		scanGroupClient: scanGroupClient,
	}
}

// Init by
func (s *Service) Init(config []byte) error {
	if config == nil || string(config) == "" {
		return errors.New("load balancer address required in Coordinator Init")
	}
	s.loadBalancerAddr = string(config)
	s.dispatcherClient = dispatcher.New()
	log.Printf("Initializing Coordinator Service\n")
	go s.connectClient()
	return nil
}

func (s *Service) connectClient() {
	err := retrier.RetryUntil(func() error {
		log.Printf("connecting to load balancer: %s\n", s.loadBalancerAddr)
		return s.dispatcherClient.Init([]byte(s.loadBalancerAddr))
	}, time.Minute*5, time.Millisecond*250)

	if err == nil {
		atomic.AddInt32(&s.connected, 1)
		log.Printf("Connected to dispatcher service\n")
		return
	}
	log.Printf("Unable to connect to dispatcher service\n")
}

func (s *Service) isConnected() bool {
	return atomic.LoadInt32(&s.connected) == 1
}

// StartGroup initializes state system if they do not exist, or updates with scan group details
func (s *Service) StartGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	log.Printf("Attempting to start group: %d\n", scanGroupID)
	if !s.isConnected() {
		log.Printf("Not connected to dispatcher..")
		return ErrNoDispatcherConnected
	}

	log.Printf("Getting scan group via client: %v %d\n", userContext, scanGroupID)
	oid, group, err := s.scanGroupClient.Get(ctx, userContext, scanGroupID)
	if err != nil {
		log.Printf("error getting group: %v\n", err)
		return err
	}

	log.Printf("Got scan group %d\n", scanGroupID)
	if oid != userContext.GetOrgID() {
		return am.ErrOrgIDMismatch
	}

	if group.Paused {
		return ErrScanGroupPaused
	}

	log.Printf("Getting group status for %d\n", scanGroupID)
	exists, _, err := s.state.GroupStatus(ctx, userContext, scanGroupID)
	if err != nil {
		return errors.Wrap(err, "failed to get group status")
	}

	if !exists {
		// update/create configuration
		log.Printf("Updating configuration for %d\n", scanGroupID)
		if err := s.state.Put(ctx, userContext, group); err != nil {
			return errors.Wrap(err, "failed to put new group")
		}
	}

	wantModules := true
	log.Printf("Getting group from state for %d\n", scanGroupID)
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
	log.Printf("Setting Start in state for %d\n", scanGroupID)
	if err := s.state.Start(ctx, userContext, group.GroupID); err != nil {
		return errors.Wrap(err, "failed to start group")
	}
	log.Printf("Dispatching in state for %d\n", scanGroupID)
	return s.dispatcherClient.PushAddresses(ctx, userContext, scanGroupID)
}
