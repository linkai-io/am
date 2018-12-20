package organization

import (
	"context"
	"io"
	"time"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/pkg/errors"

	"github.com/linkai-io/am/am"
	service "github.com/linkai-io/am/protocservices/organization"
	"google.golang.org/grpc"
)

type Client struct {
	client         service.OrganizationClient
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 10)}
}

func (c *Client) Init(config []byte) error {
	balancer := grpc.RoundRobin(grpclb.NewResolver(&grpclb.Options{
		Address: string(config),
	}))

	conn, err := grpc.Dial(am.OrganizationServiceKey, grpc.WithInsecure(), grpc.WithBalancer(balancer))
	if err != nil {
		return err
	}

	c.client = service.NewOrganizationClient(conn)
	return nil
}

func (c *Client) get(ctx context.Context, userContext am.UserContext, in *service.OrgRequest) (oid int, org *am.Organization, err error) {
	var resp *service.OrgResponse

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Get(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to get organizations from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, nil, err
	}

	return int(resp.GetOrgID()), convert.OrganizationToDomain(resp.Org), nil
}

func (c *Client) Get(ctx context.Context, userContext am.UserContext, orgName string) (oid int, org *am.Organization, err error) {
	in := &service.OrgRequest{
		By:          service.OrgRequest_ORGNAME,
		UserContext: convert.DomainToUserContext(userContext),
		OrgName:     orgName,
	}
	return c.get(ctx, userContext, in)
}

func (c *Client) GetByCID(ctx context.Context, userContext am.UserContext, orgCID string) (oid int, org *am.Organization, err error) {
	in := &service.OrgRequest{
		By:          service.OrgRequest_ORGCID,
		UserContext: convert.DomainToUserContext(userContext),
		OrgCID:      orgCID,
	}
	return c.get(ctx, userContext, in)
}

func (c *Client) GetByID(ctx context.Context, userContext am.UserContext, orgID int) (oid int, org *am.Organization, err error) {
	in := &service.OrgRequest{
		By:          service.OrgRequest_ORGID,
		UserContext: convert.DomainToUserContext(userContext),
		OrgID:       int32(orgID),
	}
	return c.get(ctx, userContext, in)
}

func (c *Client) GetByAppClientID(ctx context.Context, userContext am.UserContext, orgAppClientID string) (oid int, org *am.Organization, err error) {
	in := &service.OrgRequest{
		By:             service.OrgRequest_ORGCLIENTAPPID,
		UserContext:    convert.DomainToUserContext(userContext),
		OrgClientAppID: orgAppClientID,
	}
	return c.get(ctx, userContext, in)
}

func (c *Client) List(ctx context.Context, userContext am.UserContext, filter *am.OrgFilter) ([]*am.Organization, error) {
	var resp service.Organization_ListClient
	var err error

	in := &service.OrgListRequest{
		UserContext: convert.DomainToUserContext(userContext),
		OrgFilter:   convert.DomainToOrgFilter(filter),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error
		resp, retryErr = c.client.List(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to list organizations from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return nil, err
	}

	orgs := make([]*am.Organization, 0)
	for {
		org, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
		orgs = append(orgs, convert.OrganizationToDomain(org.Org))
	}
	return orgs, nil
}

func (c *Client) Create(ctx context.Context, userContext am.UserContext, org *am.Organization, userCID string) (oid int, uid int, ocid string, ucid string, err error) {
	var resp *service.OrgCreatedResponse
	in := &service.CreateOrgRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Org:         convert.DomainToOrganization(org),
		UserCID:     userCID,
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Create(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to create organizations from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, 0, "", "", err
	}
	return int(resp.GetOrgID()), int(resp.GetUserID()), resp.GetOrgCID(), resp.GetUserCID(), nil
}

func (c *Client) Update(ctx context.Context, userContext am.UserContext, org *am.Organization) (oid int, err error) {
	var resp *service.OrgUpdatedResponse

	in := &service.UpdateOrgRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Org:         convert.DomainToOrganization(org),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Update(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to update organizations from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, err
	}
	return int(resp.GetOrgID()), nil
}

func (c *Client) Delete(ctx context.Context, userContext am.UserContext, orgID int) (oid int, err error) {
	var resp *service.OrgDeletedResponse
	in := &service.DeleteOrgRequest{
		UserContext: convert.DomainToUserContext(userContext),
		OrgID:       int32(orgID),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Delete(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to delete organizations from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, err
	}
	return int(resp.GetOrgID()), nil
}
