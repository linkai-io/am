package protoc

import (
	"errors"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/metrics/load"
	"github.com/linkai-io/am/protocservices/address"
	context "golang.org/x/net/context"
)

var (
	ErrOrgIDNonMatch      = errors.New("error organization id's did not match")
	ErrMissingUserContext = errors.New("error request was missing user context")
	ErrNoAddressesSent    = errors.New("error no addresses were sent")
	ErrNilUserContext     = errors.New("error empty user context")
)

type AddressProtocService struct {
	as       am.AddressService
	reporter *load.RateReporter
}

func New(implementation am.AddressService, reporter *load.RateReporter) *AddressProtocService {
	return &AddressProtocService{as: implementation, reporter: reporter}
}

func (s *AddressProtocService) Get(in *address.AddressesRequest, stream address.Address_GetServer) error {
	s.reporter.Increment(1)
	defer s.reporter.Increment(-1)
	filter := convert.AddressFilterToDomain(in.Filter)

	oid, addresses, err := s.as.Get(stream.Context(), convert.UserContextToDomain(in.UserContext), filter)
	if err != nil {
		return err
	}

	for _, a := range addresses {
		if oid != a.OrgID {
			return ErrOrgIDNonMatch
		}

		if err := stream.Send(&address.AddressesResponse{Addresses: convert.DomainToAddress(a)}); err != nil {
			return err
		}
	}
	return nil
}

func (s *AddressProtocService) Update(ctx context.Context, in *address.UpdateAddressRequest) (*address.UpdateAddressesResponse, error) {
	var oid int
	var count int
	var err error
	var userContext am.UserContext
	s.reporter.Increment(1)
	addresses := make(map[string]*am.ScanGroupAddress, len(in.Address))
	for k, v := range in.Address {
		addresses[k] = convert.AddressToDomain(v)
	}
	userContext = convert.UserContextToDomain(in.UserContext)
	oid, count, err = s.as.Update(ctx, userContext, addresses)
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &address.UpdateAddressesResponse{OrgID: int32(oid), Count: int32(count)}, nil
}

func (s *AddressProtocService) Delete(ctx context.Context, in *address.DeleteAddressesRequest) (*address.DeleteAddressesResponse, error) {
	s.reporter.Increment(1)
	oid, err := s.as.Delete(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID), in.AddressIDs)
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &address.DeleteAddressesResponse{OrgID: int32(oid)}, nil
}

func (s *AddressProtocService) Count(ctx context.Context, in *address.CountAddressesRequest) (*address.CountAddressesResponse, error) {
	s.reporter.Increment(1)
	oid, count, err := s.as.Count(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &address.CountAddressesResponse{OrgID: int32(oid), GroupID: in.GroupID, Count: int32(count)}, nil
}

func (s *AddressProtocService) Ignore(ctx context.Context, in *address.IgnoreAddressesRequest) (*address.IgnoreAddressesResponse, error) {
	s.reporter.Increment(1)
	oid, err := s.as.Ignore(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID), in.AddressIDs, in.IgnoreValue)
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}
	return &address.IgnoreAddressesResponse{OrgID: int32(oid)}, nil
}
