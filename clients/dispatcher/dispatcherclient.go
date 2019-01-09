package dispatcher

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/pkg/errors"

	service "github.com/linkai-io/am/protocservices/dispatcher"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

type Client struct {
	client         service.DispatcherClient
	conn           *grpc.ClientConn
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 20)}
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.DialContext(context.Background(), "srv://consul/"+am.DispatcherServiceKey, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = service.NewDispatcherClient(conn)
	return nil
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.defaultTimeout = timeout
}

func (c *Client) PushAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	in := &service.PushRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(scanGroupID),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	return retrier.RetryIfNot(func() error {
		var retryErr error

		_, retryErr = c.client.PushAddresses(ctxDeadline, in)
		return errors.Wrap(retryErr, "failed to push addresses")
	}, "rpc error: code = Unavailable desc")

}
