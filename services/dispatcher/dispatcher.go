package dispatcher

import (
	"context"
	"encoding/json"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
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

// Config ...
type Config struct {
	DispatcherID string `json:"dispatcher_id"`
}

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
	status           int32
	defaultDuration  time.Duration
	config           *Config
	sgClient         am.ScanGroupService
	addressClient    am.AddressService
	moduleClients    map[am.ModuleType]am.ModuleService
	state            state.Stater
	pushCh           chan *pushDetails
	closeCh          chan struct{}
	completedCh      chan *am.ScanGroupAddress
	activeGroupCount int32
	activeAddrCount  int32
}

// New for coordinating the work of workers
func New(sgClient am.ScanGroupService, addrClient am.AddressService, modClients map[am.ModuleType]am.ModuleService, stater state.Stater) *Service {
	return &Service{
		defaultDuration: time.Duration(-5) * time.Minute,
		state:           stater,
		sgClient:        sgClient,
		addressClient:   addrClient,
		moduleClients:   modClients,
		pushCh:          make(chan *pushDetails),
		closeCh:         make(chan struct{}),
		completedCh:     make(chan *am.ScanGroupAddress),
	}
}

// Init this dispatcher and register it with coordinator
func (s *Service) Init(config []byte) error {
	go s.groupListener()
	go s.debug()
	return nil
}

func (s *Service) parseConfig(data []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (s *Service) debug() {
	t := time.NewTicker(time.Second * 10)
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
			log.Info().Int32("groups", s.GetActiveGroups()).Int32("addrs", s.GetActiveAddresses()).Msg("stats")
		}
	}
}

// PushAddresses to state
func (s *Service) PushAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	log.Info().Msgf("pushing details for %d", scanGroupID)
	s.pushCh <- &pushDetails{userContext: userContext, scanGroupID: scanGroupID}
	log.Info().Msgf("pushed details for %dn", scanGroupID)
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	close(s.closeCh)
	return nil
}

func (s *Service) groupListener() {
	log.Info().Msg("Listening for new scan groups to be pushed...")
	for {
		select {
		case <-s.closeCh:
			log.Info().Msg("Closing down...")
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

			// TODO: do smart calculation on size of scan group addresses
			start := time.Now()
			then := start.Add(s.defaultDuration).UnixNano()
			filter := newFilter(details.userContext, details.scanGroupID, then)

			// for now, one batcher per scan group id, todo move to own service.
			batcher := NewBatcher(details.userContext, s.addressClient, 10)
			batcher.Init()

			s.IncActiveGroups()

			scangroup, err := s.getScanGroup(ctx, details.userContext, details.scanGroupID)
			if err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("not starting analysis of group")
				goto DONE
			}

			// push addresses to state
			log.Ctx(ctx).Info().Msg("Pushing addresses to state")
			for {
				_, addrs, err := s.addressClient.Get(ctx, details.userContext, filter)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msg("error getting addresses from client")
					goto DONE
				}

				if addrs == nil || len(addrs) == 0 {
					log.Ctx(ctx).Info().Msg("no addresses matched address service filter")
					break
				}
				numAddrs := len(addrs)

				// get last addressid and update start for filter.
				filter.Start = addrs[numAddrs-1].AddressID
				log.Ctx(ctx).Info().Int("addresses", numAddrs).Msg("Putting in state")

				if err := s.state.PutAddresses(ctx, details.userContext, details.scanGroupID, addrs); err != nil {
					log.Ctx(ctx).Error().Err(err).Int64("filter.Start", filter.Start).Msg("error pushing addresses")
					goto DONE
				}
			}

			log.Ctx(ctx).Info().Msg("Push addresses complete")

			if err := s.analyzeAddresses(ctx, details.userContext, batcher, scangroup); err != nil {
				log.Ctx(ctx).Error().Err(err).Msg("error analyzing addresses")
			}

		DONE:
			groupLog.Info().Float64("group_analysis_time_seconds", time.Now().Sub(start).Seconds()).Msg("Group analysis complete")
			batcher.Done()

			if err := s.state.Stop(ctx, details.userContext, details.scanGroupID); err != nil {
				groupLog.Error().Err(err).Msg("error stopping group")
			} else {
				groupLog.Info().Msg("stopped group")
			}
			s.DecActiveGroups()
		} // end switch
	} // end for
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

// analyzeAddresses iterate over addresses from state and call analyzeAddress for each address returned. Use a worker pool
// allowing up to NSModule.RequestsPerSecond worker pool to work concurrently.
func (s *Service) analyzeAddresses(ctx context.Context, userContext am.UserContext, batcher *Batcher, scangroup *am.ScanGroup) error {

	for {
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
					s.IncActiveAddresses()
					s.processAddress(details)
					s.DecActiveAddresses()
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
		OrgID:            userContext.GetOrgID(),
		GroupID:          scanGroupID,
		Start:            0,
		Limit:            1000,
		WithLastSeenTime: true,
		SinceSeenTime:    then,
	}
}
