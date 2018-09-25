package address

import (
	"context"
	"io"
	"log"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	service "github.com/linkai-io/am/protocservices/address"
	"github.com/linkai-io/am/protocservices/prototypes"
	"google.golang.org/grpc"
)

type Client struct {
	client service.AddressClient
}

func New() *Client {
	return &Client{}
}

func (c *Client) Init(config []byte) error {
	balancer := grpc.RoundRobin(grpclb.NewResolver(&grpclb.Options{
		Address: string(config),
	}))

	conn, err := grpc.Dial(am.AddressServiceKey, grpc.WithInsecure(), grpc.WithBalancer(balancer))
	if err != nil {
		return err
	}

	c.client = service.NewAddressClient(conn)

	return nil
}

func (c *Client) Get(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, addresses []*am.ScanGroupAddress, err error) {
	var resp service.Address_GetClient

	in := &service.AddressesRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToAddressFilter(filter),
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Get(ctx, in)
		return err
	})

	if err != nil {
		return 0, nil, err
	}

	addresses = make([]*am.ScanGroupAddress, 0)
	for {
		addr, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return 0, nil, err
		}
		addresses = append(addresses, convert.AddressToDomain(addr.Addresses))
		oid = int(addr.GetOrgID())
	}
	return oid, addresses, nil
}

func (c *Client) Update(ctx context.Context, userContext am.UserContext, addresses []*am.ScanGroupAddress) (oid int, count int, err error) {
	var resp *service.UpdateAddressesResponse

	protoAddresses := make([]*prototypes.AddressData, 0)

	for i := 0; i < len(addresses); i++ {
		if addresses[i] == nil {
			log.Printf("nil address\n")
			continue
		}
		protoAddresses = append(protoAddresses, convert.DomainToAddress(addresses[i]))
	}

	in := &service.UpdateAddressRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Address:     protoAddresses,
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Update(ctx, in)
		return err
	})

	if err != nil {
		return 0, 0, err
	}

	return int(resp.GetOrgID()), int(resp.GetCount()), nil
}

func (c *Client) Count(ctx context.Context, userContext am.UserContext, groupID int) (oid int, count int, err error) {
	var resp *service.CountAddressesResponse

	in := &service.CountAddressesRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(groupID),
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Count(ctx, in)
		return err
	})

	if err != nil {
		return 0, 0, err
	}

	return int(resp.GetOrgID()), int(resp.Count), nil
}

func (c *Client) Delete(ctx context.Context, userContext am.UserContext, groupID int, addressIDs []int64) (oid int, err error) {
	var resp *service.DeleteAddressesResponse
	in := &service.DeleteAddressesRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(groupID),
		AddressIDs:  addressIDs,
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Delete(ctx, in)
		return err
	})

	if err != nil {
		return 0, err
	}

	return int(resp.GetOrgID()), nil
}
