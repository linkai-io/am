package dispatcher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/services/coordinator/state"
)

type Config struct {
	DispatcherID string `json:"dispatcher_id"`
}

// Service for coordinating the lifecycle of workers
type Service struct {
	env               string
	region            string
	config            *Config
	addressClient     am.AddressService
	coordinatorClient am.CoordinatorService
	state             state.Stater
}

// New for coordinating the work of workers
func New(env, region string, addressClient am.AddressService, coordinatorClient am.CoordinatorService, stater state.Stater) *Service {
	s := &Service{state: stater, addressClient: addressClient, coordinatorClient: coordinatorClient, env: env, region: region}
	return s
}

func (s *Service) Init(config []byte) error {
	var err error

	s.config, err = s.parseConfig(config)
	if err != nil {
		return err
	}
	ctx := context.Background()
	return s.coordinatorClient.Register(ctx, s.config.DispatcherID)
}

func (s *Service) parseConfig(data []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// push addresses to state
func (s *Service) pushAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
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

	// push addresses to state & input queue
	for {
		_, addrs, err := s.addressClient.Get(ctx, userContext, filter)
		if err != nil {
			return err
		}

		if len(addrs) == 0 {
			break
		}
		// get last addressid and update start for filter.
		filter.Start = addrs[len(addrs)-1].AddressID
	}

	return nil
}
