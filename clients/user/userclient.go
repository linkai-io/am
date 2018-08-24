package user

import (
	"context"
	"io"

	"github.com/linkai-io/am/pkg/convert"

	"google.golang.org/grpc"
	"github.com/linkai-io/am/am"
	service "github.com/linkai-io/am/protocservices/user"
)

type Client struct {
	client service.UserServiceClient
}

func New() *Client {
	return &Client{}
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.Dial(string(config))
	if err != nil {
		return err
	}
	c.client = service.NewUserServiceClient(conn)
	return nil
}

func (c *Client) get(ctx context.Context, userContext am.UserContext, in *service.UserRequest) (oid int, user *am.User, err error) {
	resp, err := c.client.Get(ctx, in)
	if err != nil {
		return 0, nil, err
	}
	return int(resp.GetOrgID()), convert.UserToDomain(resp.User), nil
}

func (c *Client) Get(ctx context.Context, userContext am.UserContext, userID int) (oid int, user *am.User, err error) {
	in := &service.UserRequest{
		UserContext: convert.DomainToUserContext(userContext),
		By:          service.UserRequest_USERID,
		UserID:      int32(userID),
	}
	return c.get(ctx, userContext, in)
}

func (c *Client) GetByCID(ctx context.Context, userContext am.UserContext, userCID string) (oid int, user *am.User, err error) {
	in := &service.UserRequest{
		UserContext: convert.DomainToUserContext(userContext),
		By:          service.UserRequest_USERCID,
		UserCID:     userCID,
	}
	return c.get(ctx, userContext, in)
}

func (c *Client) List(ctx context.Context, userContext am.UserContext, filter *am.UserFilter) (oid int, users []*am.User, err error) {
	in := &service.UserListRequest{
		UserContext: convert.DomainToUserContext(userContext),
		UserFilter:  convert.DomainToUserFilter(filter),
	}
	resp, err := c.client.List(ctx, in)
	if err != nil {
		return 0, nil, err
	}

	users = make([]*am.User, 0)
	for {
		userResp, err := resp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return 0, nil, err
		}
		users = append(users, convert.UserToDomain(userResp.User))
	}
	return 0, users, nil
}

func (c *Client) Create(ctx context.Context, userContext am.UserContext, user *am.User) (oid int, uid int, ucid string, err error) {
	in := &service.CreateUserRequest{
		UserContext: convert.DomainToUserContext(userContext),
		User:        convert.DomainToUser(user),
	}
	resp, err := c.client.Create(ctx, in)
	if err != nil {
		return 0, 0, "", err
	}
	return int(resp.GetOrgID()), int(resp.GetUserID()), resp.GetUserCID(), nil
}

func (c *Client) Update(ctx context.Context, userContext am.UserContext, user *am.User, userID int) (oid int, uid int, err error) {
	in := &service.UpdateUserRequest{
		UserContext: convert.DomainToUserContext(userContext),
		User:        convert.DomainToUser(user),
		UserID:      int32(userID),
	}
	resp, err := c.client.Update(ctx, in)
	if err != nil {
		return 0, 0, err
	}
	return int(resp.GetOrgID()), int(resp.GetUserID()), nil
}

func (c *Client) Delete(ctx context.Context, userContext am.UserContext, userID int) (oid int, err error) {
	in := &service.DeleteUserRequest{
		UserContext: convert.DomainToUserContext(userContext),
		UserID:      int32(userID),
	}
	resp, err := c.client.Delete(ctx, in)
	if err != nil {
		return 0, err
	}
	return int(resp.GetOrgID()), nil
}
