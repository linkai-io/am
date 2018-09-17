package dispatcher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/services/dispatcher/state"
	"github.com/pkg/errors"
)

// Config ...
type Config struct {
	DispatcherID string `json:"dispatcher_id"`
}

// Service for dispatching and handling responses from worker modules
type Service struct {
	config            *Config
	addressClient     am.AddressService
	coordinatorClient am.CoordinatorService
	moduleClients     map[am.ModuleType]am.ModuleService
	state             state.Stater
}

// New for coordinating the work of workers
func New(addrClient am.AddressService, coordClient am.CoordinatorService, modClients map[am.ModuleType]am.ModuleService, stater state.Stater) *Service {
	return &Service{
		state:             stater,
		addressClient:     addrClient,
		coordinatorClient: coordClient,
		moduleClients:     modClients,
	}
}

// Init this dispatcher and register it with coordinator
func (s *Service) Init(config []byte) error {
	var err error

	s.config, err = s.parseConfig(config)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return s.coordinatorClient.Register(ctx, ":50056", s.config.DispatcherID)
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
	now := time.Now()
	// TODO: do smart calculation on size of scan group addresses
	then := now.Add(time.Duration(-4) * time.Hour).UnixNano()
	filter := &am.ScanGroupAddressFilter{
		OrgID:               userContext.GetOrgID(),
		GroupID:             scanGroupID,
		Start:               0,
		Limit:               1000,
		WithLastScannedTime: true,
		SinceScannedTime:    then,
		WithIgnored:         true,
	}
	count := 0
	// push addresses to state
	for {
		_, addrs, err := s.addressClient.Get(ctx, userContext, filter)
		if err != nil {
			return err
		}
		count += len(addrs)
		if len(addrs) == 0 {
			break
		}
		// get last addressid and update start for filter.
		filter.Start = addrs[len(addrs)-1].AddressID

		if err := s.state.PutAddresses(ctx, userContext, scanGroupID, addrs); err != nil {
			log.Printf("error pushing addresses last addr: %d for scangroup %d: %s\n", filter.Start, scanGroupID, err)
			return err
		}
	}

	log.Printf("push addresses for %d complete.\n", scanGroupID)

	for {
		addrMap, err := s.state.GetAddresses(ctx, userContext, scanGroupID, 1000)
		if err != nil {
			return errors.Wrap(err, "error getting addresses")
		}

		if len(addrMap) == 0 {
			log.Printf("no more addresses for %d\n", scanGroupID)
			break
		}

		// TODO: add concurrency here
		for _, addr := range addrMap {
			s.moduleClients[am.NSModule].Analyze(ctx, addr)
		}
	}

	return nil
}
