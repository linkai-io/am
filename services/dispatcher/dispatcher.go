package dispatcher

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/services/dispatcher/state"
	"github.com/pkg/errors"
)

// DispatcherStatus for determining if we are started/stopped
type DispatcherStatus int32

const (
	Unknown DispatcherStatus = 0
	Started DispatcherStatus = 1
	Stopped DispatcherStatus = 2
)

type pushDetails struct {
	userContext am.UserContext
	scanGroupID int
}

// taskDetails contains all details needed to execute an analysis task against an am.ScanGroupAddress
type taskDetails struct {
	ctx         context.Context      // for logging
	completeCh  chan struct{}        // signaling address analysis complete
	scangroup   *am.ScanGroup        // scan group configuration
	address     *am.ScanGroupAddress // address to execute task against
	userContext am.UserContext       // user context
	batcher     *Batcher             // for storing the updated results of the analysis
	logger      zerolog.Logger       // logger specific to this task
}

// Service for dispatching and handling responses from worker modules
type Service struct {
	status           int32                              // service status
	groupCache       *cache.ScanGroupSubscriber         // listen for scan group updates from cache (deleted/paused)
	defaultDuration  time.Duration                      // filter used to extract addresses from address service
	sgClient         am.ScanGroupService                // scangroup service connection
	addressClient    am.AddressService                  // address service connection
	moduleClients    map[am.ModuleType]am.ModuleService // map of module service connections
	state            state.Stater                       // state connection
	pushCh           chan *pushDetails                  // channel for pushing groups
	closeCh          chan struct{}                      // channel for closing down service
	activeGroupCount int32                              // number of concurrent active groups
	activeAddrCount  int32                              // number of active addresses being analyzed by this dispatcher
	statGroups       *am.ScanGroupsStats                // updated stats of each group being analyzed by this dispatcher
}

// New for dispatching groups to be analyzed to the modules
func New(sgClient am.ScanGroupService, addrClient am.AddressService, modClients map[am.ModuleType]am.ModuleService, stater state.Stater) *Service {
	return &Service{
		defaultDuration: time.Duration(-5) * time.Minute,
		groupCache:      cache.NewScanGroupSubscriber(context.Background(), stater),
		state:           stater,
		sgClient:        sgClient,
		addressClient:   addrClient,
		moduleClients:   modClients,
		pushCh:          make(chan *pushDetails),
		closeCh:         make(chan struct{}),
		statGroups:      am.NewScanGroupsStats(),
	}
}

// Init this dispatcher and register it with coordinator
func (s *Service) Init(config []byte) error {
	go s.groupListener()
	go s.groupMonitor()
	return nil
}

// groupMonitor monitors status of groups and pushes updated group stats to the scan group service.
func (s *Service) groupMonitor() {
	t := time.NewTicker(time.Second * 30)
	stackTicker := time.NewTicker(time.Minute * 15)
	defer t.Stop()
	defer stackTicker.Stop()

	for {
		select {
		case <-s.closeCh:
			return
		case <-stackTicker.C:
			buf := make([]byte, 1<<20)
			stacklen := runtime.Stack(buf, true)
			log.Printf("*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
		case <-t.C:

			log.Info().Int32("groups", s.GetActiveGroups()).Int32("addrs", s.GetActiveAddresses()).Msg("updating group stats")

			timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			for _, stats := range s.statGroups.Groups() {
				_, err := s.sgClient.UpdateStats(timeoutCtx, stats.UserContext, stats)
				if err != nil {
					log.Error().Err(err).Int("GroupID", stats.GroupID).Int("OrgID", stats.OrgID).Msg("failed to update stats for group")
				}
			}
			cancel()
		}
	}
}

// PushAddresses to state
func (s *Service) PushAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	log.Info().Msgf("pushing details for %d", scanGroupID)
	if s.GetActiveAddresses() > 2000 {
		log.Warn().Int("GroupID", scanGroupID).Int32("active_addresses", s.GetActiveAddresses()).Msg("active addresses too high, not handling this group")
		if err := s.state.Stop(ctx, userContext, scanGroupID); err != nil {
			log.Error().Err(err).Msg("error stopping group")
		} else {
			log.Info().Msg("stopped group due to too many active addresses")
		}
		return nil
	}
	s.pushCh <- &pushDetails{userContext: userContext, scanGroupID: scanGroupID}
	log.Info().Msgf("pushed details for %d", scanGroupID)
	return nil
}

// Stop the service
func (s *Service) Stop(ctx context.Context) error {
	close(s.closeCh)
	return nil
}

// groupListener listens for new group messages coming in, and ensures after completion that the
// group is stopped in state, and any relevant counters are stopped.
func (s *Service) groupListener() {
	log.Info().Msg("Listening for new scan groups to be pushed...")
	for {
		select {
		case <-s.closeCh:
			log.Info().Msg("Closing down...")
			for _, group := range s.statGroups.Groups() {
				s.stopGroup(context.Background(), group.UserContext, group.GroupID)
			}
			return
		case details := <-s.pushCh:
			ctx := context.Background()
			groupLog := log.With().
				Int("UserID", details.userContext.GetUserID()).
				Int("GroupID", details.scanGroupID).
				Int("OrgID", details.userContext.GetOrgID()).
				Str("TraceID", details.userContext.GetTraceID()).Logger()
			ctx = groupLog.WithContext(ctx)

			log.Ctx(ctx).Info().Msg("Starting Analysis")

			start := time.Now()
			s.runGroup(ctx, details, start)
			log.Ctx(ctx).Info().Float64("group_analysis_time_seconds", time.Now().Sub(start).Seconds()).Msg("Group analysis complete")

			s.stopGroup(ctx, details.userContext, details.scanGroupID)
			s.DecActiveGroups()
		}
	}
}

func (s *Service) stopGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) {

	s.statGroups.SetComplete(scanGroupID)
	stats := s.statGroups.GetGroup(scanGroupID)
	if stats != nil {
		_, err := s.sgClient.UpdateStats(ctx, userContext, stats)
		if err != nil {
			log.Error().Err(err).Int("GroupID", stats.GroupID).Int("OrgID", stats.OrgID).Msg("failed to update stats for group")
		}
	}

	if err := s.state.Stop(ctx, userContext, scanGroupID); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error stopping group")
	} else {
		log.Ctx(ctx).Info().Msg("stopped group")
	}
	s.statGroups.DeleteGroup(scanGroupID)
}

// runGroup sets up the scan group batcher to push results to the address service, and extracts all
// addresses that haven't been run for defaultDuration time (5 min). Those addresses are pushed
// on to the cache state. After all addresses have been pushed, analyzeAddresses begins.
func (s *Service) runGroup(ctx context.Context, details *pushDetails, start time.Time) {

	s.statGroups.AddGroup(details.userContext, details.userContext.GetOrgID(), details.scanGroupID)
	// for now, one batcher per scan group id, todo move to own service.
	batcher := NewBatcher(details.userContext, s.addressClient, 50)
	batcher.Init()
	defer batcher.Done()

	// TODO: do smart calculation on size of scan group addresses
	then := start.Add(s.defaultDuration).UnixNano()
	filter := newFilter(details.userContext, details.scanGroupID, then)

	s.IncActiveGroups()

	scangroup, err := s.getScanGroup(ctx, details.userContext, details.scanGroupID)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("not starting analysis of group")
		return
	}

	// push addresses to state
	total := 0
	log.Ctx(ctx).Info().Msg("Pushing addresses to state")
	for {
		log.Ctx(ctx).Info().Msgf("Getting addresses that match filter: %#v", filter)
		_, addrs, err := s.addressClient.Get(ctx, details.userContext, filter)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("error getting addresses from client")
			return
		}

		if addrs == nil || len(addrs) == 0 {
			log.Ctx(ctx).Info().Msg("no addresses matched address service filter")
			break
		}
		numAddrs := len(addrs)
		total += numAddrs

		// get last addressid and update start for filter.
		filter.Start = addrs[numAddrs-1].AddressID
		log.Ctx(ctx).Info().Int("addresses", numAddrs).Msg("Putting in state")

		if err := s.state.PutAddresses(ctx, details.userContext, details.scanGroupID, addrs); err != nil {
			log.Ctx(ctx).Error().Err(err).Int64("filter.Start", filter.Start).Msg("error pushing addresses")
			return
		}
	}

	log.Ctx(ctx).Info().Msg("Push addresses complete")
	s.statGroups.SetBatchSize(details.scanGroupID, int32(total))

	if err := s.analyzeAddresses(ctx, details.userContext, batcher, scangroup); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("error analyzing addresses")
	}
}

func (s *Service) getScanGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) (*am.ScanGroup, error) {
	oid, scangroup, err := s.sgClient.Get(ctx, userContext, scanGroupID)
	if err != nil {
		return nil, err
	}

	if oid != userContext.GetOrgID() {
		return nil, am.ErrOrgIDMismatch
	}

	return scangroup, nil
}

// shouldStopGroup inspects the updated group state from cache to see if it's been paused/deleted
func (s *Service) shouldStopGroup(orgID, groupID int) bool {
	newGroup, err := s.groupCache.GetGroupByIDs(orgID, groupID)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get group from cache to check if paused/deleted")
		return false
	}

	if newGroup.Paused == true || newGroup.Deleted == true {
		return true
	}
	return false
}

// analyzeAddresses iterate over addresses from state and call analyzeAddress for each address returned. Use a worker pool
// allowing up to NSModule.RequestsPerSecond worker pool to work concurrently.
func (s *Service) analyzeAddresses(ctx context.Context, userContext am.UserContext, batcher *Batcher, scangroup *am.ScanGroup) error {
	for {
		if s.shouldStopGroup(scangroup.OrgID, scangroup.GroupID) {
			return errors.New("group was paused or deleted during analysis")
		}

		addrMap, err := s.state.PopAddresses(ctx, userContext, scangroup.GroupID, 1000)
		if err != nil {
			return errors.Wrap(err, "error getting addresses")
		}
		numAddrs := len(addrMap)

		if numAddrs == 0 {
			log.Ctx(ctx).Info().Msg("no more addresses from work queue")
			break
		}

		log.Ctx(ctx).Info().Int("address_count", len(addrMap)).Msg("popped from state")
		rps := int(scangroup.ModuleConfigurations.NSModule.RequestsPerSecond)

		if numAddrs < rps {
			rps = numAddrs
		}

		pool := workerpool.New(rps)
		log.Ctx(ctx).Info().Int("worker_pool", rps).Msg("created for processing dispatcher tasks")

		for _, addr := range addrMap {
			analysisAddr := addr

			task := func(details *taskDetails) func() {
				return func() {
					group, err := s.groupCache.GetGroupByIDs(details.scangroup.OrgID, details.scangroup.GroupID)
					if err == nil {
						if group.Paused || group.Deleted {
							return
						}
					} else {
						log.Ctx(ctx).Warn().Err(err).Msg("failed to get group from cache during process tasks, continuing")
					}
					s.statGroups.IncActive(details.scangroup.GroupID, 1)
					s.IncActiveAddresses()
					s.processAddress(details)
					s.DecActiveAddresses()
					s.statGroups.IncActive(details.scangroup.GroupID, -1)
				}
			}

			details := &taskDetails{
				ctx:         ctx,
				scangroup:   scangroup,
				address:     analysisAddr,
				userContext: userContext,
				batcher:     batcher,
			}

			pool.Submit(task(details))
		}
		pool.StopWait()
	}
	return nil
}

// processAddress for a given task, running the analysis of the address, and adding the final results to our batcher.
func (s *Service) processAddress(details *taskDetails) {
	ctx := details.ctx
	updatedAddress, err := s.analyzeAddress(ctx, details.userContext, details.scangroup.GroupID, details.address)
	if err != nil {
		// TODO: need to figure out how to handle not losing hosts, but also not scanning forever if
		// they are always problematic.
		log.Ctx(ctx).Error().Err(err).Str("ip", details.address.IPAddress).Str("host", details.address.HostAddress).Msg("failed to analyze address")
		s.updateOriginal(details.batcher, details.address)
		return
	}

	updatedAddress.LastSeenTime = time.Now().UnixNano()
	details.batcher.Add(updatedAddress)

	// this happens iff input_list/manual hosts only have one of ip or host
	if details.address.AddressHash != updatedAddress.AddressHash {
		s.updateOriginal(details.batcher, details.address)
	}
}

// updateOriginal since we can not update the original input hosts (or manually added)
// but we don't want our last seen check to keep re-adding the hosts for scans.
func (s *Service) updateOriginal(batcher *Batcher, originalAddress *am.ScanGroupAddress) {
	switch originalAddress.DiscoveredBy {
	case am.DiscoveryNSInputList, am.DiscoveryNSManual:
		now := time.Now().UnixNano()
		originalAddress.LastScannedTime = now
		originalAddress.LastSeenTime = now
		batcher.Add(originalAddress)
	}
}

// analyzeAddress analyzes ns records, then brute forces, then web systems. (TODO: add bigdata / other modules)
func (s *Service) analyzeAddress(ctx context.Context, userContext am.UserContext, scanGroupID int, address *am.ScanGroupAddress) (*am.ScanGroupAddress, error) {
	logger := log.Ctx(ctx).With().Int64("AddressID", address.AddressID).Str("IPAddress", address.IPAddress).Str("HostAddress", address.HostAddress).Logger()
	ctx = logger.WithContext(ctx)

	log.Ctx(ctx).Info().Msg("analyzing ns records")
	updatedAddress, err := s.moduleAnalysis(ctx, userContext, s.moduleClients[am.NSModule], scanGroupID, address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to analyze using ns module")
	}

	log.Ctx(ctx).Info().Msg("brute forcing")
	updatedAddress, err = s.moduleAnalysis(ctx, userContext, s.moduleClients[am.BruteModule], scanGroupID, updatedAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to analyze using brute module")
	}

	log.Ctx(ctx).Info().Msg("bigquery ct subdomain lookup")
	updatedAddress, err = s.moduleAnalysis(ctx, userContext, s.moduleClients[am.BigDataCTSubdomainModule], scanGroupID, updatedAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to analyze using brute module")
	}

	// web analysis requires a valid AddressID
	if updatedAddress.AddressID == 0 {
		log.Ctx(ctx).Info().Msg("skipping web analysis until AddressID exists")
		return updatedAddress, nil
	}

	log.Ctx(ctx).Info().Msg("analyzing web systems")
	updatedAddress, err = s.moduleAnalysis(ctx, userContext, s.moduleClients[am.WebModule], scanGroupID, updatedAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to analyze using web module")
	}

	return updatedAddress, nil
}

func (s *Service) moduleAnalysis(ctx context.Context, userContext am.UserContext, module am.ModuleService, scanGroupID int, address *am.ScanGroupAddress) (*am.ScanGroupAddress, error) {
	updatedAddress, possibleNewAddrs, err := module.Analyze(ctx, userContext, address)
	if err != nil {
		return nil, err
	}

	if len(possibleNewAddrs) == 0 {
		return updatedAddress, nil
	}

	newAddrs, err := s.state.FilterNew(ctx, userContext.GetOrgID(), scanGroupID, possibleNewAddrs)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("testing state to determine new address failed")
	}

	if len(newAddrs) > 0 {
		if err := s.state.PutAddressMap(ctx, userContext, scanGroupID, newAddrs); err != nil {
			log.Ctx(ctx).Error().Err(err).Int("address_count", len(newAddrs)).Msg("failed to put in state")
		}
	}
	return updatedAddress, nil
}

func (s *Service) IncActiveGroups() {
	atomic.AddInt32(&s.activeGroupCount, 1)
}

func (s *Service) DecActiveGroups() {
	atomic.AddInt32(&s.activeGroupCount, -1)
}

func (s *Service) GetActiveGroups() int32 {
	return atomic.LoadInt32(&s.activeGroupCount)
}

func (s *Service) IncActiveAddresses() {
	atomic.AddInt32(&s.activeAddrCount, 1)
}

func (s *Service) DecActiveAddresses() {
	atomic.AddInt32(&s.activeAddrCount, -1)
}

func (s *Service) GetActiveAddresses() int32 {
	return atomic.LoadInt32(&s.activeAddrCount)
}

func newFilter(userContext am.UserContext, scanGroupID int, then int64) *am.ScanGroupAddressFilter {
	return &am.ScanGroupAddressFilter{
		OrgID:                  userContext.GetOrgID(),
		GroupID:                scanGroupID,
		Start:                  0,
		Limit:                  1000,
		WithBeforeLastSeenTime: true,
		BeforeSeenTime:         then,
	}
}
