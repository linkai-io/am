package dispatcher

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"

	service "github.com/linkai-io/am/protocservices/dispatcher"
	"google.golang.org/grpc"
)

type Client struct {
	client service.DispatcherClient
}

func New() *Client {
	return &Client{}
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.Dial(string(config), grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.client = service.NewDispatcherClient(conn)
	return nil
}

func (c *Client) PushAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	in := &service.PushRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(scanGroupID),
	}

	_, err := c.client.PushAddresses(ctx, in)
	return err
}
