package dispatcher

import (
	"context"
	"time"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/pkg/errors"

	service "github.com/linkai-io/am/protocservices/dispatcher"
	"google.golang.org/grpc"
)

type Client struct {
	client         service.DispatcherClient
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 20)}
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

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	return retrier.RetryUnless(func() error {
		var retryErr error

		_, retryErr = c.client.PushAddresses(ctxDeadline, in)
		return errors.Wrap(retryErr, "failed to push addresses")
	}, am.ErrUserNotAuthorized)

}
