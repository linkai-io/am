package mock

import (
	"context"

	"github.com/linkai-io/am/am"
)

type UserService struct {
	InitFn      func(config []byte) error
	InitInvoked bool

	GetFn      func(ctx context.Context, userContext am.UserContext, userEmail string) (oid int, user *am.User, err error)
	GetInvoked bool

	GetWithOrgIDFn      func(ctx context.Context, userContext am.UserContext, orgID int, userCID string) (oid int, org *am.User, err error)
	GetWithOrgIDInvoked bool

	GetByIDFn      func(ctx context.Context, userContext am.UserContext, userID int) (oid int, user *am.User, err error)
	GetByIDInvoked bool

	GetByCIDFn      func(ctx context.Context, userContext am.UserContext, userCID string) (oid int, user *am.User, err error)
	GetByCIDInvoked bool

	ListFn      func(ctx context.Context, userContext am.UserContext, filter *am.UserFilter) (oid int, users []*am.User, err error)
	ListInvoked bool

	CreateFn      func(ctx context.Context, userContext am.UserContext, user *am.User) (oid int, uid int, ucid string, err error)
	CreateInvoked bool

	UpdateFn      func(ctx context.Context, userContext am.UserContext, user *am.User, userID int) (oid int, uid int, err error)
	UpdateInvoked bool

	DeleteFn      func(ctx context.Context, userContext am.UserContext, userID int) (oid int, err error)
	DeleteInvoked bool
}

func (c *UserService) Init(config []byte) error {
	c.InitInvoked = true
	return c.InitFn(config)
}

func (c *UserService) Get(ctx context.Context, userContext am.UserContext, userEmail string) (oid int, user *am.User, err error) {
	c.GetInvoked = true
	return c.GetFn(ctx, userContext, userEmail)
}

func (c *UserService) GetByCID(ctx context.Context, userContext am.UserContext, orgCID string) (oid int, user *am.User, err error) {
	c.GetByCIDInvoked = true
	return c.GetByCIDFn(ctx, userContext, orgCID)
}

func (c *UserService) GetWithOrgID(ctx context.Context, userContext am.UserContext, orgID int, userCID string) (oid int, user *am.User, err error) {
	c.GetWithOrgIDInvoked = true
	return c.GetWithOrgIDFn(ctx, userContext, orgID, userCID)
}

func (c *UserService) GetByID(ctx context.Context, userContext am.UserContext, userID int) (oid int, user *am.User, err error) {
	c.GetByIDInvoked = true
	return c.GetByIDFn(ctx, userContext, userID)
}

func (c *UserService) List(ctx context.Context, userContext am.UserContext, filter *am.UserFilter) (oid int, users []*am.User, err error) {
	c.ListInvoked = true
	return c.ListFn(ctx, userContext, filter)
}

func (c *UserService) Create(ctx context.Context, userContext am.UserContext, user *am.User) (oid int, uid int, ucid string, err error) {
	c.CreateInvoked = true
	return c.CreateFn(ctx, userContext, user)
}

func (c *UserService) Update(ctx context.Context, userContext am.UserContext, user *am.User, userID int) (oid int, uid int, err error) {
	c.UpdateInvoked = true
	return c.UpdateFn(ctx, userContext, user, userID)
}

func (c *UserService) Delete(ctx context.Context, userContext am.UserContext, userID int) (oid int, err error) {
	c.DeleteInvoked = true
	return c.DeleteFn(ctx, userContext, userID)
}
