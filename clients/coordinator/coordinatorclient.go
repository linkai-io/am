package coordinator

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"

	service "github.com/linkai-io/am/protocservices/coordinator"
	"google.golang.org/grpc"
)

type Client struct {
	client service.CoordinatorClient
}

func New() *Client {
	return &Client{}
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.Dial(string(config), grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.client = service.NewCoordinatorClient(conn)
	return nil
}

func (c *Client) StartGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	in := &service.StartGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(scanGroupID),
	}

	_, err := c.client.StartGroup(ctx, in)
	return err
}

func (c *Client) Register(ctx context.Context, address, dispatcherID string) error {
	in := &service.RegisterRequest{
		DispatcherAddr: address,
		DispatcherID:   dispatcherID,
	}

	_, err := c.client.Register(ctx, in)
	return err
}
