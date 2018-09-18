package dispatcher

import (
	"context"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"

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
	balancer := grpc.RoundRobin(grpclb.NewResolver(&grpclb.Options{
		Address: string(config),
	}))

	conn, err := grpc.Dial(am.DispatcherServiceKey, grpc.WithInsecure(), grpc.WithBalancer(balancer))
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

	return retrier.Retry(func() error {
		var err error
		_, err = c.client.PushAddresses(ctx, in)
		return err
	})

}
