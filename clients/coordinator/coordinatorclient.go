package coordinator

import (
	"context"
	"time"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/pkg/errors"

	service "github.com/linkai-io/am/protocservices/coordinator"
	"google.golang.org/grpc"
)

type Client struct {
	client         service.CoordinatorClient
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 10)}
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
	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	return retrier.Retry(func() error {
		var retryErr error

		_, retryErr = c.client.StartGroup(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to start group from coordinator client")
	})
}
