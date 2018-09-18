package scangroup

import (
	"context"
	"io"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	service "github.com/linkai-io/am/protocservices/scangroup"
	"google.golang.org/grpc"
)

type Client struct {
	client service.ScanGroupClient
}

func New() *Client {
	return &Client{}
}

func (c *Client) Init(config []byte) error {
	balancer := grpc.RoundRobin(grpclb.NewResolver(&grpclb.Options{
		Address: string(config),
	}))

	conn, err := grpc.Dial(am.ScanGroupServiceKey, grpc.WithInsecure(), grpc.WithBalancer(balancer))
	if err != nil {
		return err
	}

	c.client = service.NewScanGroupClient(conn)
	return nil
}

func (c *Client) get(ctx context.Context, userContext am.UserContext, in *service.GroupRequest) (oid int, group *am.ScanGroup, err error) {
	var resp *service.GroupResponse

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Get(ctx, in)
		return err
	})

	if err != nil {
		return 0, nil, err
	}
	return int(resp.GetOrgID()), convert.ScanGroupToDomain(resp.GetGroup()), nil
}

func (c *Client) Get(ctx context.Context, userContext am.UserContext, groupID int) (oid int, group *am.ScanGroup, err error) {
	in := &service.GroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		By:          service.GroupRequest_GROUPID,
		GroupID:     int32(groupID),
	}
	return c.get(ctx, userContext, in)
}

func (c *Client) GetByName(ctx context.Context, userContext am.UserContext, groupName string) (oid int, group *am.ScanGroup, err error) {
	in := &service.GroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		By:          service.GroupRequest_GROUPNAME,
		GroupName:   groupName,
	}
	return c.get(ctx, userContext, in)
}

func (c *Client) Groups(ctx context.Context, userContext am.UserContext) (oid int, groups []*am.ScanGroup, err error) {
	var stream service.ScanGroup_GroupsClient

	in := &service.GroupsRequest{
		UserContext: convert.DomainToUserContext(userContext),
	}

	err = retrier.Retry(func() error {
		var err error
		stream, err = c.client.Groups(ctx, in)
		return err
	})

	if err != nil {
		return 0, nil, err
	}

	groups = make([]*am.ScanGroup, 0)
	for {
		group, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return 0, nil, err
		}
		groups = append(groups, convert.ScanGroupToDomain(group.GetGroup()))
		oid = int(group.GetOrgID())
	}

	return oid, groups, nil
}

func (c *Client) Create(ctx context.Context, userContext am.UserContext, newGroup *am.ScanGroup) (oid int, gid int, err error) {
	var resp *service.GroupCreatedResponse

	in := &service.NewGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Group:       convert.DomainToScanGroup(newGroup),
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Create(ctx, in)
		return err
	})

	if err != nil {
		return 0, 0, err
	}

	return int(resp.GetOrgID()), int(resp.GetGroupID()), nil
}

func (c *Client) Update(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) (oid int, gid int, err error) {
	var resp *service.GroupUpdatedResponse

	in := &service.UpdateGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Group:       convert.DomainToScanGroup(group),
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Update(ctx, in)
		return err
	})

	if err != nil {
		return 0, 0, err
	}

	return int(resp.GetOrgID()), int(resp.GetGroupID()), nil
}

func (c *Client) Delete(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	var resp *service.GroupDeletedResponse

	in := &service.DeleteGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(groupID),
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Delete(ctx, in)
		return err
	})

	if err != nil {
		return 0, 0, err
	}

	return int(resp.GetOrgID()), int(resp.GetGroupID()), nil
}

func (c *Client) Pause(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	var resp *service.GroupPausedResponse

	in := &service.PauseGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(groupID),
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Pause(ctx, in)
		return err
	})

	if err != nil {
		return 0, 0, err
	}

	return int(resp.GetOrgID()), int(resp.GetGroupID()), nil
}

func (c *Client) Resume(ctx context.Context, userContext am.UserContext, groupID int) (oid int, gid int, err error) {
	var resp *service.GroupResumedResponse

	in := &service.ResumeGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(groupID),
	}

	err = retrier.Retry(func() error {
		var err error
		resp, err = c.client.Resume(ctx, in)
		return err
	})

	if err != nil {
		return 0, 0, err
	}

	return int(resp.GetOrgID()), int(resp.GetGroupID()), nil
}
