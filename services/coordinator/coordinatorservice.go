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
	orgClient        am.OrganizationService

	connected        int32
	dispatcherClient am.DispatcherService

	stopCh chan struct{}
}

// New returns
func New(state state.Stater, dispatcherClient am.DispatcherService, orgClient am.OrganizationService, scanGroupClient am.ScanGroupService, systemOrgID, systemUserID int) *Service {
	return &Service{
		state:            state,
		orgClient:        orgClient,
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

func (s *Service) getOrg(ctx context.Context, systemContext am.UserContext, orgID int) (*am.Organization, error) {
	_, org, err := s.orgClient.GetByID(ctx, systemContext, orgID)
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (s *Service) startGroups() {
	ctx := context.Background()
	systemContext := &am.UserContextData{OrgID: s.systemOrgID, UserID: s.systemUserID, TraceID: createID()}

	groups, err := s.scanGroupClient.AllGroups(ctx, systemContext, &am.ScanGroupFilter{Filters: &am.FilterType{}})
	if err != nil {
		log.Error().Err(err).Msg("failed to get groups")
		return
	}

	for _, group := range groups {
		org, err := s.getOrg(ctx, systemContext, group.OrgID)
		if err != nil {
			log.Error().Err(err).Int("OrgID", group.OrgID).Msg("failed to get org cid for org id")
			continue
		}

		proxyUserContext := &am.UserContextData{OrgID: group.OrgID, OrgCID: org.OrgCID, SubscriptionID: org.SubscriptionID, UserID: group.CreatedByID, TraceID: createID()}
		log.Info().Int("GroupID", group.GroupID).Msg("starting group")
		if err := s.StartGroup(ctx, proxyUserContext, group.GroupID); err != nil {
			if err != ErrScanGroupAlreadyStarted {
				log.Warn().Err(err).Msg("failed to start group")
			}
		}
	}
}

// shouldUpdateGroup checks if the group should be updated (modified time check) even if it is running
func (s *Service) shouldUpdateGroup(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) error {
	wantModules := true
	cachedGroup, err := s.state.GetGroup(ctx, userContext.GetOrgID(), group.GroupID, wantModules)
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
	return nil
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

	groupLog.Info().Bool("exists", exists).Int("status", int(status)).Msg("Getting group via scangroupclient")
	oid, group, err := s.scanGroupClient.Get(ctx, userContext, scanGroupID)
	if err != nil {
		return errors.Wrap(err, "err with scan group client")
	}

	if status == am.GroupStarted || group.Paused || group.Deleted {
		if err := s.shouldUpdateGroup(ctx, userContext, group); err != nil {
			groupLog.Error().Err(err).Msg("shouldUpdate group failed")
			return err
		}
		groupLog.Info().Msg("Group already started")
		return ErrScanGroupAlreadyStarted
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

	if err := s.shouldUpdateGroup(ctx, userContext, group); err != nil {
		groupLog.Error().Err(err).Msg("failed during shouldUpdateGroup check")
		return errors.Wrap(err, "failed to start group")
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

// StopGroup checks the group status in our state system. If it's already stopped, we are done. Otherwise,
// it extracts group from scangroup client to create a proxyUserContext.
func (s *Service) StopGroup(ctx context.Context, userContext am.UserContext, orgID, scanGroupID int) (string, error) {
	if userContext.GetOrgID() != s.systemOrgID && userContext.GetUserID() != s.systemUserID {
		return "", am.ErrUserNotAuthorized
	}

	// need the org id to actually look it up from state.
	exists, status, err := s.state.GroupStatus(ctx, &am.UserContextData{OrgID: orgID}, scanGroupID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get group status")
	}

	log.Info().Bool("exists", exists).Int("status", int(status)).Msg("stop group got group status")

	if status == am.GroupStopped {
		return "group already stopped", nil
	}

	if exists {
		if err := s.state.Stop(ctx, &am.UserContextData{OrgID: orgID}, scanGroupID); err != nil {
			log.Error().Err(err).Int("OrgID", orgID).Int("GroupID", scanGroupID).Msg("failed to set state as stopped")
			return "", err
		}
		return "group stopped", nil
	}

	return "group not found", nil
}

// stopGroup gets the updated group data and then deletes the old group from state, replacing it
// with the updated, stopped group
func (s *Service) stopGroup(ctx context.Context, userContext am.UserContext, exists bool, group *am.ScanGroup) (string, error) {

	log.Info().Msg("Getting group via scangroupclient")
	_, pausedGroup, err := s.scanGroupClient.Get(ctx, userContext, group.GroupID)
	if err != nil {
		return "", errors.Wrap(err, "err with scan group client")
	}

	if err := s.state.Delete(ctx, userContext, pausedGroup); err != nil {
		return "", errors.Wrap(err, "failed to delete group")
	}

	// update/create configuration
	if err := s.state.Put(ctx, userContext, pausedGroup); err != nil {
		return "", errors.Wrap(err, "failed to update/put group")
	}
	return "", nil
}

func createID() string {
	id, err := uuid.NewV4()
	if err != nil {
		log.Warn().Err(err).Msg("unable to generate new traceid")
		return "empty_coordinator_trace_id"
	}
	return id.String()
}
