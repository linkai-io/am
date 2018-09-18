package organization

import (
	"context"
	"io"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"

	"github.com/linkai-io/am/am"
	service "github.com/linkai-io/am/protocservices/organization"
	"google.golang.org/grpc"
)

type Client struct {
	client service.OrganizationClient
}

func New() *Client {
	return &Client{}
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
	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Get(ctx, in)
		return err
	})

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

func (c *Client) List(ctx context.Context, userContext am.UserContext, filter *am.OrgFilter) ([]*am.Organization, error) {
	var resp service.Organization_ListClient
	in := &service.OrgListRequest{
		UserContext: convert.DomainToUserContext(userContext),
		OrgFilter:   convert.DomainToOrgFilter(filter),
	}

	err := retrier.Retry(func() error {
		var err error
		resp, err = c.client.List(ctx, in)
		return err
	})

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

func (c *Client) Create(ctx context.Context, userContext am.UserContext, org *am.Organization) (oid int, uid int, ocid string, ucid string, err error) {
	var resp *service.OrgCreatedResponse
	in := &service.CreateOrgRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Org:         convert.DomainToOrganization(org),
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Create(ctx, in)
		return err
	})

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

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Update(ctx, in)
		return err
	})

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

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Delete(ctx, in)
		return err
	})

	if err != nil {
		return 0, err
	}
	return int(resp.GetOrgID()), nil
}
