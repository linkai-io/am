package bigdata

import (
	"context"
	"time"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/pkg/errors"

	"github.com/linkai-io/am/am"
	service "github.com/linkai-io/am/protocservices/bigdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

type Client struct {
	client         service.BigDataClient
	conn           *grpc.ClientConn
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 60)}
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.DialContext(context.Background(), "srv://consul/"+am.BigDataServiceKey, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = service.NewBigDataClient(conn)
	return nil
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.defaultTimeout = timeout
}

func (c *Client) GetCT(ctx context.Context, userContext am.UserContext, etld string) (time.Time, map[string]*am.CTRecord, error) {
	var resp *service.GetCTResponse
	var err error
	var emptyTS time.Time

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	in := &service.GetCTRequest{
		UserContext: convert.DomainToUserContext(userContext),
		ETLD:        etld,
	}

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.GetCT(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get ct records from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return emptyTS, nil, err
	}
	return time.Unix(0, resp.Time), convert.CTRecordsToDomain(resp.Records), nil
}

func (c *Client) AddCT(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, ctRecords map[string]*am.CTRecord) error {
	var err error

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	in := &service.AddCTRequest{
		UserContext: convert.DomainToUserContext(userContext),
		ETLD:        etld,
		QueryTime:   queryTime.UnixNano(),
		Records:     convert.DomainToCTRecords(ctRecords),
	}

	err = retrier.RetryIfNot(func() error {
		_, retryErr := c.client.AddCT(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to add ct records from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteCT(ctx context.Context, userContext am.UserContext, etld string) error {
	var err error

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	in := &service.DeleteCTRequest{
		UserContext: convert.DomainToUserContext(userContext),
		ETLD:        etld,
	}

	err = retrier.RetryIfNot(func() error {
		_, retryErr := c.client.DeleteCT(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to delete ct records from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return err
	}
	return nil
}
