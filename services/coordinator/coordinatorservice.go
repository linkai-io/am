package coordinator

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/pkg/retrier"

	"github.com/pkg/errors"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/dispatcher"
	"github.com/linkai-io/am/services/coordinator/state"
)

var (
	ErrScanGroupPaused         = errors.New("scan group is currently paused")
	ErrScanGroupAlreadyStarted = errors.New("scan group is already running")
	ErrNoDispatcherConnected   = errors.New("service unavailable, no dispatchers connected")
)

// Service for coordinating all scans
type Service struct {
	systemOrgID      int
	systemUserID     int
	loadBalancerAddr string
	state            state.Stater
	scanGroupClient  am.ScanGroupService

	connected        int32
	dispatcherClient am.DispatcherService

	stopCh chan struct{}
}

// New returns
func New(state state.Stater, scanGroupClient am.ScanGroupService, systemOrgID, systemUserID int) *Service {
	return &Service{
		state:           state,
		scanGroupClient: scanGroupClient,
		systemOrgID:     systemOrgID,
		systemUserID:    systemUserID,
		stopCh:          make(chan struct{}),
	}
}

// Init by
func (s *Service) Init(config []byte) error {
	if config == nil || string(config) == "" {
		return errors.New("load balancer address required in Coordinator Init")
	}
	s.loadBalancerAddr = string(config)
	s.dispatcherClient = dispatcher.New()
	log.Info().Msg("Initializing Coordinator Service")
	go s.connectClient()
	go s.poller()
	return nil
}

func (s *Service) connectClient() {
	err := retrier.RetryUntil(func() error {
		log.Info().Str("load balancer", s.loadBalancerAddr).Msg("connecting to load balancer")
		return s.dispatcherClient.Init([]byte(s.loadBalancerAddr))
	}, time.Minute*5, time.Millisecond*250)

	if err == nil {
		atomic.AddInt32(&s.connected, 1)
		log.Info().Msg("Connected to dispatcher service\n")
		return
	}
	log.Warn().Msg("Unable to connect to dispatcher service\n")
}

func (s *Service) isConnected() bool {
	return atomic.LoadInt32(&s.connected) == 1
}

func (s *Service) poller() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Info().Msg("Scanning to start groups")
			s.startGroups()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Service) startGroups() {
	ctx := context.Background()
	userContext := &am.UserContextData{OrgID: s.systemOrgID, UserID: s.systemUserID, TraceID: createID()}

	groups, err := s.scanGroupClient.AllGroups(ctx, userContext, &am.ScanGroupFilter{WithPaused: true, PausedValue: false})
	if err != nil {
		log.Error().Err(err).Msg("failed to get groups")
		return
	}

	for _, group := range groups {
		proxyUserContext := &am.UserContextData{OrgID: group.OrgID, UserID: group.CreatedBy, TraceID: createID()}
		log.Info().Int("GroupID", group.GroupID).Msg("starting group")
		if err := s.StartGroup(ctx, proxyUserContext, group.GroupID); err != nil {
			if err != ErrScanGroupAlreadyStarted {
				log.Warn().Err(err).Msg("failed to start group")
			}
		}
	}
}

// StartGroup initializes state system if they do not exist, or updates with scan group details
func (s *Service) StartGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	groupLog := log.With().
		Int("UserID", userContext.GetUserID()).
		Int("GroupID", scanGroupID).
		Int("OrgID", userContext.GetOrgID()).
		Str("TraceID", userContext.GetTraceID()).Logger()

	groupLog.Info().Msg("Attempting to start group")
	if !s.isConnected() {
		groupLog.Warn().Msg("Not connected to dispatcher")
		return ErrNoDispatcherConnected
	}

	groupLog.Info().Msg("Getting group status")
	exists, status, err := s.state.GroupStatus(ctx, userContext, scanGroupID)
	if err != nil {
		return errors.Wrap(err, "failed to get group status")
	}

	if status == am.GroupStarted {
		return ErrScanGroupAlreadyStarted
	}

	groupLog.Info().Msg("Getting group via scangroupclient")
	oid, group, err := s.scanGroupClient.Get(ctx, userContext, scanGroupID)
	if err != nil {
		return errors.Wrap(err, "err with scan group client")
	}

	groupLog.Info().Msg("Got scan group from client")
	if oid != userContext.GetOrgID() {
		return am.ErrOrgIDMismatch
	}

	if group.Paused {
		return ErrScanGroupPaused
	}

	if !exists {
		// update/create configuration
		groupLog.Info().Msg("Updating configuration for scangroup")
		if err := s.state.Put(ctx, userContext, group); err != nil {
			return errors.Wrap(err, "failed to put new group")
		}
	}

	wantModules := true
	groupLog.Info().Msg("Getting scangroup from state")
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

	groupLog.Info().Msg("Setting Start in state for scangroup")
	if err := s.state.Start(ctx, userContext, group.GroupID); err != nil {
		return errors.Wrap(err, "failed to start group")
	}

	groupLog.Info().Msg("Dispatching scangroup")
	return s.dispatcherClient.PushAddresses(ctx, userContext, scanGroupID)
}

func createID() string {
	id, err := uuid.NewV4()
	if err != nil {
		log.Warn().Err(err).Msg("unable to generate new traceid")
		return "empty_coordinator_trace_id"
	}
	return id.String()
}
