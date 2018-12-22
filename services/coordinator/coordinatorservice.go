package coordinator

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"

	"github.com/pkg/errors"

	"github.com/linkai-io/am/am"
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
func New(state state.Stater, dispatcherClient am.DispatcherService, scanGroupClient am.ScanGroupService, systemOrgID, systemUserID int) *Service {
	return &Service{
		state:            state,
		scanGroupClient:  scanGroupClient,
		dispatcherClient: dispatcherClient,
		systemOrgID:      systemOrgID,
		systemUserID:     systemUserID,
		stopCh:           make(chan struct{}),
	}
}

// Init by
func (s *Service) Init(config []byte) error {
	go s.poller()
	return nil
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
		proxyUserContext := &am.UserContextData{OrgID: group.OrgID, UserID: group.CreatedByID, TraceID: createID()}
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
		// TODO: empty work queue if it exists
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

	err = s.dispatcherClient.PushAddresses(ctx, userContext, scanGroupID)
	if err != nil {
		if stopErr := s.state.Stop(ctx, userContext, scanGroupID); stopErr != nil {
			groupLog.Error().Err(stopErr).Msg("failed to set state to stopped after push addresses failed")
		} else {
			groupLog.Warn().Msg("stopped group due to push address failure")
		}
	}

	return err
}

func createID() string {
	id, err := uuid.NewV4()
	if err != nil {
		log.Warn().Err(err).Msg("unable to generate new traceid")
		return "empty_coordinator_trace_id"
	}
	return id.String()
}
