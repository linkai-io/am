package webdata

import (
	"context"
	"time"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/pkg/errors"

	"github.com/linkai-io/am/am"
	service "github.com/linkai-io/am/protocservices/webdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

type Client struct {
	client         service.WebDataClient
	conn           *grpc.ClientConn
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 60)}
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.DialContext(context.Background(), "srv://consul/"+am.WebDataServiceKey, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = service.NewWebDataClient(conn)
	return nil
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.defaultTimeout = timeout
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

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Add(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get add records from client")
	}, "rpc error: code = Unavailable desc")

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

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.GetResponses(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get get responses from client")
	}, "rpc error: code = Unavailable desc")

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

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.GetCertificates(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get ct records from client")
	}, "rpc error: code = Unavailable desc")

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

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.GetSnapshots(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get snapshots from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, nil, err
	}
	return int(resp.OrgID), convert.WebSnapshotsToDomain(resp.Snapshots), nil
}

func (c *Client) GetURLList(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, []*am.URLListResponse, error) {
	var resp *service.GetURLListResponse
	var err error

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	in := &service.GetURLListRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToWebResponseFilter(filter),
	}

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.GetURLList(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get url list from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, nil, err
	}
	return int(resp.OrgID), convert.URLListsToDomain(resp.GetURLList()), nil
}

func (c *Client) GetDomainDependency(ctx context.Context, userContext am.UserContext, filter *am.WebResponseFilter) (int, *am.WebDomainDependency, error) {
	var resp *service.GetDomainDependencyResponse
	var err error

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	in := &service.GetDomainDependencyRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToWebResponseFilter(filter),
	}

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.GetDomainDependency(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get domain deps from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, nil, err
	}
	return int(resp.OrgID), convert.WebDomainDependencyToDomain(resp.Dependency), nil
}

func (c *Client) OrgStats(ctx context.Context, userContext am.UserContext) (oid int, orgStats []*am.ScanGroupWebDataStats, err error) {
	var resp *service.OrgStatsResponse
	oid = userContext.GetOrgID()

	in := &service.OrgStatsRequest{
		UserContext: convert.DomainToUserContext(userContext),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.OrgStats(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to get webdata org stats from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, nil, err
	}
	return int(resp.GetOrgID()), convert.ScanGroupsWebDataStatsToDomain(resp.GetGroupStats()), nil
}

func (c *Client) GroupStats(ctx context.Context, userContext am.UserContext, groupID int) (oid int, groupStats *am.ScanGroupWebDataStats, err error) {
	var resp *service.GroupStatsResponse
	oid = userContext.GetOrgID()

	in := &service.GroupStatsRequest{
		UserContext: convert.DomainToUserContext(userContext),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.GroupStats(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to get webdata group stats from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, nil, err
	}
	return int(resp.GetOrgID()), convert.ScanGroupWebDataStatsToDomain(resp.GetGroupStats()), nil
}

func (c *Client) Archive(ctx context.Context, userContext am.UserContext, group *am.ScanGroup, archiveTime time.Time) (oid int, count int, err error) {
	var resp *service.WebArchivedResponse
	oid = userContext.GetOrgID()

	in := &service.ArchiveWebRequest{
		UserContext: convert.DomainToUserContext(userContext),
		ScanGroup:   convert.DomainToScanGroup(group),
		ArchiveTime: archiveTime.UnixNano(),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Archive(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to get web archive from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, 0, err
	}
	return int(resp.GetOrgID()), int(resp.Count), nil
}
