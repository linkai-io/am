package webdata

import (
	"context"
	"time"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/pkg/errors"

	"github.com/linkai-io/am/am"
	service "github.com/linkai-io/am/protocservices/webdata"
	"google.golang.org/grpc"
)

type Client struct {
	client         service.WebDataClient
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 60)}
}

func (c *Client) Init(config []byte) error {
	balancer := grpc.RoundRobin(grpclb.NewResolver(&grpclb.Options{
		Address: string(config),
	}))

	conn, err := grpc.Dial(am.WebDataServiceKey, grpc.WithInsecure(), grpc.WithBalancer(balancer))
	if err != nil {
		return err
	}

	c.client = service.NewWebDataClient(conn)
	return nil
}

func (c *Client) Add(ctx context.Context, userContext am.UserContext, webData *am.WebData) (int, error) {
	var resp *service.AddedResponse
	var err error

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	in := &service.AddRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Data:        convert.DomainToWebData(webData),
	}

	err = retrier.Retry(func() error {
		var retryErr error

		resp, retryErr = c.client.Add(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get add records from client")
	})

	if err != nil {
		return 0, err
	}
	return int(resp.OrgID), nil
}

func (c *Client) GetResponses(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, []*am.HTTPResponse, error) {
	var resp *service.GetResponsesResponse
	var err error

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	in := &service.GetResponsesRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToWebResponseFilter(filter),
	}

	err = retrier.Retry(func() error {
		var retryErr error

		resp, retryErr = c.client.GetResponses(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get ct records from client")
	})

	if err != nil {
		return 0, nil, err
	}
	return int(resp.OrgID), convert.HTTPResponsesToDomain(resp.Responses), nil
}

func (c *Client) GetCertificates(ctx context.Context, userContext am.UserContext, filter *am.WebCertificateFilter) (int, []*am.WebCertificate, error) {
	var resp *service.GetCertificatesResponse
	var err error

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	in := &service.GetCertificatesRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToWebCertificateFilter(filter),
	}

	err = retrier.Retry(func() error {
		var retryErr error

		resp, retryErr = c.client.GetCertificates(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get ct records from client")
	})

	if err != nil {
		return 0, nil, err
	}

	return int(resp.OrgID), convert.WebCertificatesToDomain(resp.Certificates), nil
}

func (c *Client) GetSnapshots(ctx context.Context, userContext am.UserContext, filter *am.WebSnapshotFilter) (int, []*am.WebSnapshot, error) {
	var resp *service.GetSnapshotsResponse
	var err error

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	in := &service.GetSnapshotsRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToWebSnapshotFilter(filter),
	}

	err = retrier.Retry(func() error {
		var retryErr error

		resp, retryErr = c.client.GetSnapshots(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get ct records from client")
	})

	if err != nil {
		return 0, nil, err
	}
	return int(resp.OrgID), convert.WebSnapshotsToDomain(resp.Snapshots), nil
}
