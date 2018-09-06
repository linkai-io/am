package address

import (
	"context"
	"io"
	"log"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/am"
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
	conn, err := grpc.Dial(string(config), grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.client = service.NewAddressClient(conn)
	return nil
}

func (c *Client) Get(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupAddressFilter) (oid int, addresses []*am.ScanGroupAddress, err error) {

	in := &service.AddressesRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToAddressFilter(filter),
	}
	resp, err := c.client.Get(ctx, in)
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
