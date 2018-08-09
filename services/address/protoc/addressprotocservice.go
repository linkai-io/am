package protoc

import (
	"errors"
	"io"

	context "golang.org/x/net/context"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/convert"
	"gopkg.linkai.io/v1/repos/am/protocservices/address"
)

var (
	ErrOrgIDNonMatch      = errors.New("error organization id's did not match")
	ErrMissingUserContext = errors.New("error request was missing user context")
	ErrNoAddressesSent    = errors.New("error no addresses were sent")
)

type AddressProtocService struct {
	as               am.AddressService
	MaxAddressStream int32
}

func New(implementation am.AddressService) *AddressProtocService {
	return &AddressProtocService{as: implementation, MaxAddressStream: 10000}
}

func (s *AddressProtocService) Addresses(in *address.AddressesRequest, stream address.Address_AddressesServer) error {
	filter := convert.AddressFilterToDomain(in.Filter)

	oid, addresses, err := s.as.Addresses(stream.Context(), convert.UserContextToDomain(in.UserContext), filter)
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

func (s *AddressProtocService) UpdatedAddresses(stream address.Address_UpdateAddressesServer) error {
	var oid int
	var updateCount int
	var count int32
	var i int32

	addresses := make([]*am.ScanGroupAddress, s.MaxAddressStream)
	for {
		addr, err := stream.Recv()
		if err == io.EOF {
			// only do update if we have addresses
			if i != 0 {
				oid, updateCount, err = s.as.Update(stream.Context(), convert.UserContextToDomain(addr.UserContext), addresses[0:i])
				if err != nil {
					return err
				}
				count += int32(updateCount)
			}
			return stream.SendAndClose(&address.UpdateAddressesResponse{OrgID: int32(oid), Count: int32(count)})
		}
		// other error occurred
		if err != nil {
			return err
		}

		if i == s.MaxAddressStream {
			oid, updateCount, err = s.as.Update(stream.Context(), convert.UserContextToDomain(addr.UserContext), addresses)
			if err != nil {
				return err
			}
			count += int32(updateCount)
			// reset the addresses slice and current index
			i = 0
			addresses = make([]*am.ScanGroupAddress, s.MaxAddressStream)
		}
		addresses[i] = convert.AddressToDomain(addr.Address)
		i++
	}
	return nil
}

func (s *AddressProtocService) AddressCount(ctx context.Context, in *address.CountAddressesRequest) (*address.CountAddressesResponse, error) {
	oid, count, err := s.as.AddressCount(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	if err != nil {
		return nil, err
	}
	return &address.CountAddressesResponse{OrgID: int32(oid), GroupID: in.GroupID, Count: int32(count)}, nil
}
