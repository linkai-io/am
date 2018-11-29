package mock

import (
	"context"

	"github.com/linkai-io/am/am"
)

type OrganizationService struct {
	InitFn      func(config []byte) error
	InitInvoked bool

	GetFn      func(ctx context.Context, userContext am.UserContext, orgName string) (oid int, org *am.Organization, err error)
	GetInvoked bool

	GetByCIDFn      func(ctx context.Context, userContext am.UserContext, orgCID string) (oid int, org *am.Organization, err error)
	GetByCIDInvoked bool

	GetByIDFn      func(ctx context.Context, userContext am.UserContext, orgID int) (oid int, org *am.Organization, err error)
	GetByIDInvoked bool

	ListFn      func(ctx context.Context, userContext am.UserContext, filter *am.OrgFilter) (orgs []*am.Organization, err error)
	ListInvoked bool

	CreateFn      func(ctx context.Context, userContext am.UserContext, org *am.Organization) (oid int, uid int, ocid string, ucid string, err error)
	CreateInvoked bool

	UpdateFn      func(ctx context.Context, userContext am.UserContext, org *am.Organization) (oid int, err error)
	UpdateInvoked bool

	DeleteFn      func(ctx context.Context, userContext am.UserContext, orgID int) (oid int, err error)
	DeleteInvoked bool
}

func (c *OrganizationService) Init(config []byte) error {
	c.InitInvoked = true
	return c.InitFn(config)
}

func (c *OrganizationService) Get(ctx context.Context, userContext am.UserContext, orgName string) (oid int, org *am.Organization, err error) {
	c.GetInvoked = true
	return c.GetFn(ctx, userContext, orgName)
}

func (c *OrganizationService) GetByCID(ctx context.Context, userContext am.UserContext, orgCID string) (oid int, org *am.Organization, err error) {
	c.GetInvoked = true
	return c.GetByCIDFn(ctx, userContext, orgCID)
}

func (c *OrganizationService) GetByID(ctx context.Context, userContext am.UserContext, orgID int) (oid int, org *am.Organization, err error) {
	c.GetInvoked = true
	return c.GetByIDFn(ctx, userContext, orgID)
}

func (c *OrganizationService) List(ctx context.Context, userContext am.UserContext, filter *am.OrgFilter) (orgs []*am.Organization, err error) {
	c.ListInvoked = true
	return c.ListFn(ctx, userContext, filter)
}

func (c *OrganizationService) Create(ctx context.Context, userContext am.UserContext, org *am.Organization) (oid int, uid int, ocid string, ucid string, err error) {
	c.CreateInvoked = true
	return c.CreateFn(ctx, userContext, org)
}

func (c *OrganizationService) Update(ctx context.Context, userContext am.UserContext, org *am.Organization) (oid int, err error) {
	c.UpdateInvoked = true
	return c.UpdateFn(ctx, userContext, org)
}

func (c *OrganizationService) Delete(ctx context.Context, userContext am.UserContext, orgID int) (oid int, err error) {
	c.DeleteInvoked = true
	return c.DeleteFn(ctx, userContext, orgID)
}
