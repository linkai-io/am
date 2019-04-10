package coordinator

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/pkg/errors"

	service "github.com/linkai-io/am/protocservices/coordinator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

type Client struct {
	client         service.CoordinatorClient
	conn           *grpc.ClientConn
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 10)}
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.DialContext(context.Background(), "srv://consul/"+am.CoordinatorServiceKey, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = service.NewCoordinatorClient(conn)
	return nil
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.defaultTimeout = timeout
}

func (c *Client) StartGroup(ctx context.Context, userContext am.UserContext, scanGroupID int) error {

	in := &service.StartGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(scanGroupID),
	}
	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	return retrier.RetryIfNot(func() error {
		var retryErr error

		_, retryErr = c.client.StartGroup(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to start group from coordinator client")
	}, "rpc error: code = Unavailable desc")
}

func (c *Client) StopGroup(ctx context.Context, userContext am.UserContext, orgID, scanGroupID int) (string, error) {
	var message *service.GroupStoppedResponse

	in := &service.StopGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		OrgID:       int32(orgID),
		GroupID:     int32(scanGroupID),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()
	err := retrier.RetryIfNot(func() error {
		var retryErr error

		message, retryErr = c.client.StopGroup(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to stop group from coordinator client")
	}, "rpc error: code = Unavailable desc")
	if err != nil {
		return "", err
	}
	return message.Message, nil
}
