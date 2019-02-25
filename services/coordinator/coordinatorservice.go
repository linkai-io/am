package coordinator

import (
	"context"
	"sync"
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

	orgLock  sync.RWMutex
	orgCache map[int]string

	connected        int32
	dispatcherClient am.DispatcherService

	stopCh chan struct{}
}

// New returns
func New(state state.Stater, dispatcherClient am.DispatcherService, orgClient am.OrganizationService, scanGroupClient am.ScanGroupService, systemOrgID, systemUserID int) *Service {
	return &Service{
		state:            state,
		orgClient:        orgClient,
		orgCache:         make(map[int]string, 0),
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

func (s *Service) getOrgCID(ctx context.Context, systemContext am.UserContext, orgID int) (string, error) {
	s.orgLock.Lock()
	defer s.orgLock.Unlock()
	if orgCID, ok := s.orgCache[orgID]; ok {
		return orgCID, nil
	}

	_, org, err := s.orgClient.GetByID(ctx, systemContext, orgID)
	if err != nil {
		return "", err
	}

	s.orgCache[orgID] = org.OrgCID
	return org.OrgCID, nil
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
		orgCID, err := s.getOrgCID(ctx, systemContext, group.OrgID)
		if err != nil {
			log.Error().Err(err).Int("OrgID", group.OrgID).Msg("failed to get org cid for org id")
			continue
		}

		proxyUserContext := &am.UserContextData{OrgID: group.OrgID, OrgCID: orgCID, UserID: group.CreatedByID, TraceID: createID()}
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

	groupLog.Info().Msg("Getting group via scangroupclient")
	oid, group, err := s.scanGroupClient.Get(ctx, userContext, scanGroupID)
	if err != nil {
		return errors.Wrap(err, "err with scan group client")
	}

	if status == am.GroupStarted || group.Paused || group.Deleted {
		s.shouldUpdateGroup(ctx, userContext, group)
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
// it extracts group from scangroup client to create a proxyUserContext. Then call stopGroup to delete the
// old config from state, and replace it with the new, paused state after calling sgClient.Pause.
func (s *Service) StopGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) (string, error) {
	if userContext.GetOrgID() != s.systemOrgID && userContext.GetUserID() != s.systemUserID {
		return "", am.ErrUserNotAuthorized
	}

	exists, status, err := s.state.GroupStatus(ctx, userContext, scanGroupID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get group status")
	}

	if status == am.GroupStopped {
		return "group already stopped", nil
	}

	filter := &am.ScanGroupFilter{
		Filters: &am.FilterType{},
	}
	filter.Filters.AddBool("paused", true)

	groups, err := s.scanGroupClient.AllGroups(ctx, userContext, filter)
	if err != nil {
		log.Error().Err(err).Msg("failed to get groups")
		return "", err
	}

	proxyUserContext := &am.UserContextData{}
	for _, group := range groups {
		if group.GroupID != scanGroupID {
			continue
		}

		orgCID, err := s.getOrgCID(ctx, userContext, group.OrgID)
		if err != nil {
			log.Error().Err(err).Int("OrgID", group.OrgID).Msg("failed to get org cid for org id")
			continue
		}

		proxyUserContext = &am.UserContextData{OrgID: group.OrgID, OrgCID: orgCID, UserID: group.CreatedByID, TraceID: createID()}
		return s.stopGroup(ctx, proxyUserContext, exists, group)
	}
	return "group not found", nil
}

// stopGroup pauses the scan group, gets the updated group data and then deletes the old group from state, replacing it
// with the updated, stopped group
func (s *Service) stopGroup(ctx context.Context, userContext am.UserContext, exists bool, group *am.ScanGroup) (string, error) {
	log.Info().Msg("Pausing scangroup")
	if _, _, err := s.scanGroupClient.Pause(ctx, userContext, group.GroupID); err != nil {
		return "", errors.Wrap(err, "err with scan group client")
	}

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
