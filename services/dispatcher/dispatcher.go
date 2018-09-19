package dispatcher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/services/dispatcher/state"
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
	config        *Config
	addressClient am.AddressService
	moduleClients map[am.ModuleType]am.ModuleService
	state         state.Stater
	pushCh        chan *pushDetails
	closeCh       chan struct{}
}

// New for coordinating the work of workers
func New(addrClient am.AddressService, modClients map[am.ModuleType]am.ModuleService, stater state.Stater) *Service {
	return &Service{
		state:         stater,
		addressClient: addrClient,
		moduleClients: modClients,
		pushCh:        make(chan *pushDetails),
		closeCh:       make(chan struct{}),
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
	log.Printf("pushing details for %d\n", scanGroupID)
	s.pushCh <- &pushDetails{userContext: userContext, scanGroupID: scanGroupID}
	log.Printf("pushed details for %d\n", scanGroupID)
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	close(s.closeCh)
	return nil
}

func (s *Service) listener() {
	log.Printf("Listening for new scan groups to be pushed...")
	for {
	LISTEN:
		select {
		case <-s.closeCh:
			log.Printf("Closing down...\n")
			return
		case details := <-s.pushCh:
			ctx := context.Background()
			now := time.Now()
			// TODO: do smart calculation on size of scan group addresses
			then := now.Add(time.Duration(-4) * time.Hour).UnixNano()
			filter := &am.ScanGroupAddressFilter{
				OrgID:               details.userContext.GetOrgID(),
				GroupID:             details.scanGroupID,
				Start:               0,
				Limit:               1000,
				WithLastScannedTime: true,
				SinceScannedTime:    then,
				WithIgnored:         true,
			}
			count := 0
			// push addresses to state
			log.Printf("Pushing addresses to state for %d\n", details.scanGroupID)
			for {
				_, addrs, err := s.addressClient.Get(ctx, details.userContext, filter)
				if err != nil {
					log.Printf("error getting addresses from client: %s\n", err)
					goto LISTEN
				}
				count += len(addrs)
				if len(addrs) == 0 {
					break
				}
				// get last addressid and update start for filter.
				filter.Start = addrs[len(addrs)-1].AddressID
				log.Printf("Putting %d addresses in state for %d\n", len(addrs), details.scanGroupID)
				if err := s.state.PutAddresses(ctx, details.userContext, details.scanGroupID, addrs); err != nil {
					log.Printf("error pushing addresses last addr: %d for scangroup %d: %s\n", filter.Start, details.scanGroupID, err)
					goto LISTEN
				}
			}

			log.Printf("Push addresses for %d complete.\n", details.scanGroupID)

			for {
				addrMap, err := s.state.GetAddresses(ctx, details.userContext, details.scanGroupID, 1000)
				if err != nil {
					log.Printf("error getting addresses: %s\n", err)
					goto LISTEN
				}
				log.Printf("got %d addresses for %d\n", len(addrMap), details.scanGroupID)

				if len(addrMap) == 0 {
					log.Printf("no more addresses for %d\n", details.scanGroupID)
					break
				}

				// TODO: add concurrency here
				for _, addr := range addrMap {
					log.Printf("dispatching %v for ns module\n", addr)
					s.moduleClients[am.NSModule].Analyze(ctx, addr)
				}
			}
		}
	}
}
