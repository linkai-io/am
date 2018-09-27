package dispatcher

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/services/dispatcher/state"
	"github.com/pkg/errors"
)

// Config ...
type Config struct {
	DispatcherID string `json:"dispatcher_id"`
}

type pushDetails struct {
	userContext am.UserContext
	scanGroupID int
}

// Service for dispatching and handling responses from worker modules
type Service struct {
	config           *Config
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
func New(addrClient am.AddressService, modClients map[am.ModuleType]am.ModuleService, stater state.Stater) *Service {
	return &Service{
		state:         stater,
		addressClient: addrClient,
		moduleClients: modClients,
		pushCh:        make(chan *pushDetails),
		closeCh:       make(chan struct{}),
		completedCh:   make(chan *am.ScanGroupAddress),
	}
}

// Init this dispatcher and register it with coordinator
func (s *Service) Init(config []byte) error {
	go s.listener()
	return nil
}

func (s *Service) parseConfig(data []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// PushAddresses to state
func (s *Service) PushAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	log.Info().Msgf("pushing details for %d", scanGroupID)
	s.pushCh <- &pushDetails{userContext: userContext, scanGroupID: scanGroupID}
	log.Info().Msgf("pushed details for %d\n", scanGroupID)
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	close(s.closeCh)
	return nil
}

func (s *Service) listener() {
	log.Info().Msg("Listening for new scan groups to be pushed...")
	for {
		select {
		case <-s.closeCh:
			log.Info().Msg("Closing down...")
			return
		case details := <-s.pushCh:
			groupLog := log.With().
				Int("UserID", details.userContext.GetUserID()).
				Int("GroupID", details.scanGroupID).
				Int("OrgID", details.userContext.GetOrgID()).
				Str("TraceID", details.userContext.GetTraceID()).Logger()

			s.IncActiveGroups()
			ctx := context.Background()
			now := time.Now()
			// TODO: do smart calculation on size of scan group addresses
			then := now.Add(time.Duration(-4) * time.Hour).UnixNano()
			filter := &am.ScanGroupAddressFilter{
				OrgID:            details.userContext.GetOrgID(),
				GroupID:          details.scanGroupID,
				Start:            0,
				Limit:            1000,
				WithLastSeenTime: true,
				SinceSeenTime:    then,
			}

			// push addresses to state
			groupLog.Info().Msg("Pushing addresses to state")
			// for now, one per scan group id, todo move to own service.
			batcher := NewBatcher(details.userContext, s.addressClient, 10)
			batcher.Init()
			for {
				_, addrs, err := s.addressClient.Get(ctx, details.userContext, filter)
				if err != nil {
					groupLog.Error().Err(err).Msg("error getting addresses from client")
					goto DONE
				}
				if addrs == nil || len(addrs) == 0 {
					break
				}
				numAddrs := len(addrs)

				// get last addressid and update start for filter.
				filter.Start = addrs[numAddrs-1].AddressID
				groupLog.Info().Int("addresses", numAddrs).Msg("putting in state")

				if err := s.state.PutAddresses(ctx, details.userContext, details.scanGroupID, addrs); err != nil {
					groupLog.Error().Err(err).Int64("filter.Start", filter.Start).Msg("error pushing addresses")
					goto DONE
				}
			}

			groupLog.Info().Msg("Push addresses complete")

			if err := s.analyzeAddresses(ctx, details.userContext, groupLog, batcher, details.scanGroupID); err != nil {
				groupLog.Error().Err(err).Msg("error analyzing addresses")
			}

		DONE:
			batcher.Done()

			if err := s.state.Stop(ctx, details.userContext, details.scanGroupID); err != nil {
				groupLog.Error().Err(err).Msg("error stopping group")
			}
			s.DecActiveGroups()
		} // end switch
	} // end for
}

// analyzeAddresses iterate over addresses from state and call analyzeAddress for each address returned
// TODO: add concurrency here
func (s *Service) analyzeAddresses(ctx context.Context, userContext am.UserContext, groupLog zerolog.Logger, batcher *Batcher, scanGroupID int) error {
	for {
		addrMap, err := s.state.PopAddresses(ctx, userContext, scanGroupID, 1000)
		if err != nil {
			return errors.Wrap(err, "error getting addresses")
		}

		if len(addrMap) == 0 {
			groupLog.Info().Msg("no more addresses")
			break
		}

		for _, addr := range addrMap {
			s.IncActiveAddresses()
			address, err := s.analyzeAddress(ctx, userContext, groupLog, scanGroupID, addr)
			if err != nil {
				groupLog.Error().Err(err).Str("ip", addr.IPAddress).Str("host", addr.HostAddress)
			}
			address.LastScannedTime = time.Now().UnixNano()
			batcher.Add(address)
			s.DecActiveAddresses()
		}
	}
	return nil
}

// analyzeAddress
func (s *Service) analyzeAddress(ctx context.Context, userContext am.UserContext, groupLog zerolog.Logger, scanGroupID int, address *am.ScanGroupAddress) (*am.ScanGroupAddress, error) {
	groupLog.Info().Str("ip", address.IPAddress).Str("host", address.HostAddress).Msg("analyzing")
	updatedAddress, possibleNewAddrs, err := s.moduleClients[am.NSModule].Analyze(ctx, address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to analyze using ns module")
	}

	s.addNewPossibleAddresses(ctx, userContext, groupLog, scanGroupID, possibleNewAddrs)

	return updatedAddress, nil
}

func (s *Service) addNewPossibleAddresses(ctx context.Context, userContext am.UserContext, groupLog zerolog.Logger, scanGroupID int, addresses map[string]*am.ScanGroupAddress) {
	// test if newAddrs already exist in set before adding
	newAddrs, err := s.state.FilterNew(ctx, userContext.GetOrgID(), scanGroupID, addresses)
	if err != nil {
		groupLog.Error().Err(err).Msg("testing state to determine new address failed")
	}

	if len(newAddrs) > 0 {
		if err := s.state.PutAddressMap(ctx, userContext, scanGroupID, newAddrs); err != nil {
			groupLog.Error().Err(err).Int("address_count", len(newAddrs)).Msg("failed to put in state")
		}
	}
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
