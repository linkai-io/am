package dispatcher

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/services/coordinator/state"
)

// Service for coordinating the lifecycle of workers
type Service struct {
	env           string
	region        string
	addressClient am.AddressService
	state         state.Stater
}

// NewService for coordinating the work of workers
func NewService(env, region string, addressClient am.AddressService, stater state.Stater) *Service {
	s := &Service{state: stater, addressClient: addressClient, env: env, region: region}
	return s
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
