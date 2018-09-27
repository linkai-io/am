package coordinator

import (
	"context"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/rs/zerolog/log"

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
	balancer := grpc.RoundRobin(grpclb.NewResolver(&grpclb.Options{
		Address: string(config),
	}))

	conn, err := grpc.Dial(am.CoordinatorServiceKey, grpc.WithInsecure(), grpc.WithBalancer(balancer))
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

	return retrier.Retry(func() error {
		_, err := c.client.StartGroup(ctx, in)
		log.Error().Err(err).Msg("error starting group in client")
		return err
	})
}
