package address

import (
	"context"
	"fmt"
	"io"
	"log"

	balancerpb "github.com/bsm/grpclb/grpclb_balancer_v1"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	service "github.com/linkai-io/am/protocservices/address"
	"google.golang.org/grpc"
)

type Client struct {
	client service.AddressClient
}

func New() *Client {
	return &Client{}
}

func (c *Client) Init(config []byte) error {
	/*
		conn, err := grpc.Dial(string(config), grpc.WithInsecure())
		if err != nil {
			return err
		}
	*/
	cc, err := grpc.Dial(string(config), grpc.WithInsecure())
	if err != nil {
		return err
	}

	bc := balancerpb.NewLoadBalancerClient(cc)
	resp, err := bc.Servers(context.Background(), &balancerpb.ServersRequest{
		Target: "addressservice",
	})

	if err != nil {
		return err
	}

	for _, srv := range resp.Servers {
		fmt.Printf("%d\t%s\n", srv.Score, srv.Address)
	}

	c.client = service.NewAddressClient(cc)
	return nil
}

func (c *Client) Get(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, addresses []*am.ScanGroupAddress, err error) {

	in := &service.AddressesRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToAddressFilter(filter),
	}
	fmt.Printf("sending get request: %#v\n", in)
	resp, err := c.client.Get(ctx, in)
	if err != nil {
		fmt.Printf("Got get error: %#v\n", err)
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

	stream, err := c.client.Update(ctx)
	if err != nil {
		return 0, 0, err
	}

	for i := 0; i < len(addresses); i++ {
		if addresses[i] == nil {
			log.Printf("nil address\n")
			continue
		}
		in := &service.UpdateAddressRequest{
			UserContext: convert.DomainToUserContext(userContext),
			Address:     convert.DomainToAddress(addresses[i]),
		}

		if err := stream.Send(in); err != nil {
			return 0, 0, err
		}

	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		return 0, 0, err
	}

	return int(reply.GetOrgID()), int(reply.GetCount()), nil
}

func (c *Client) Count(ctx context.Context, userContext am.UserContext, groupID int) (oid int, count int, err error) {
	in := &service.CountAddressesRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(groupID),
	}
	resp, err := c.client.Count(ctx, in)
	if err != nil {
		return 0, 0, err
	}
	return int(resp.GetOrgID()), int(resp.Count), nil
}

func (c *Client) Delete(ctx context.Context, userContext am.UserContext, groupID int, addressIDs []int64) (oid int, err error) {
	in := &service.DeleteAddressesRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(groupID),
		AddressIDs:  addressIDs,
	}
	resp, err := c.client.Delete(ctx, in)
	if err != nil {
		return 0, err
	}

	return int(resp.GetOrgID()), nil
}
