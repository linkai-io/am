package scangroup

import (
	"context"
	"io"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	service "github.com/linkai-io/am/protocservices/scangroup"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

type Client struct {
	client         service.ScanGroupClient
	conn           *grpc.ClientConn
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 10)}
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.defaultTimeout = timeout
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.DialContext(context.Background(), "srv://consul/"+am.ScanGroupServiceKey, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = service.NewScanGroupClient(conn)
	return nil
}

func (c *Client) get(ctx context.Context, userContext am.UserContext, in *service.GroupRequest) (oid int, group *am.ScanGroup, err error) {
	var resp *service.GroupResponse

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Get(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get scangroup from client")
	}, "rpc error: code = Unavailable desc")

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

func (c *Client) AllGroups(ctx context.Context, userContext am.UserContext, filter *am.ScanGroupFilter) (groups []*am.ScanGroup, err error) {
	var stream service.ScanGroup_AllGroupsClient

	in := &service.AllGroupsRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToScanGroupFilter(filter),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error
		stream, retryErr = c.client.AllGroups(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to get all scangroups from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		log.Error().Err(err).Msg("UNABLE TO GET GROUPS")
		return nil, err
	}

	groups = make([]*am.ScanGroup, 0)
	for {
		group, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
		groups = append(groups, convert.ScanGroupToDomain(group.GetGroup()))
	}

	return groups, nil

}

func (c *Client) Groups(ctx context.Context, userContext am.UserContext) (oid int, groups []*am.ScanGroup, err error) {
	var stream service.ScanGroup_GroupsClient
	oid = userContext.GetOrgID()

	in := &service.GroupsRequest{
		UserContext: convert.DomainToUserContext(userContext),
	}
	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		stream, retryErr = c.client.Groups(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to get groups from scangroup client")
	}, "rpc error: code = Unavailable desc")

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

		domainGroup := group.GetGroup()
		// empty group
		if domainGroup.GetOrgID() == 0 && domainGroup.GetGroupID() == 0 {
			continue
		}

		groups = append(groups, convert.ScanGroupToDomain(domainGroup))
		if domainGroup.GetOrgID() != int32(oid) {
			log.Info().Msgf("%#v OID: %d", domainGroup, oid)
			return 0, nil, am.ErrOrgIDMismatch
		}
	}

	return oid, groups, nil
}

func (c *Client) Create(ctx context.Context, userContext am.UserContext, newGroup *am.ScanGroup) (oid int, gid int, err error) {
	var resp *service.GroupCreatedResponse

	in := &service.NewGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Group:       convert.DomainToScanGroup(newGroup),
	}
	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Create(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to create groups from scangroup client")
	}, "rpc error: code = Unavailable desc")

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

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Update(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to update group from scangroup client")
	}, "rpc error: code = Unavailable desc")

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

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Delete(ctxDeadline, in)
		cancel()
		return errors.Wrap(retryErr, "unable to delete group from scangroup client")
	}, "rpc error: code = Unavailable desc")

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

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Pause(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to pause group from scangroup client")
	}, "rpc error: code = Unavailable desc")

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

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Resume(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to resume group from scangroup client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, 0, err
	}

	return int(resp.GetOrgID()), int(resp.GetGroupID()), nil
}

func (c *Client) GroupStats(ctx context.Context, userContext am.UserContext) (oid int, stats map[int]*am.GroupStats, err error) {
	var resp *service.GroupStatsResponse

	in := &service.GroupStatsRequest{
		UserContext: convert.DomainToUserContext(userContext),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.GroupStats(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get scangroup from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, nil, err
	}
	return int(resp.GetOrgID()), convert.GroupsStatsToDomain(resp.GetStats()), nil
}

func (c *Client) UpdateStats(ctx context.Context, userContext am.UserContext, stats *am.GroupStats) (oid int, err error) {
	var resp *service.StatsUpdatedResponse

	in := &service.UpdateStatsRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Stats:       convert.DomainToGroupStats(stats),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.UpdateStats(ctxDeadline, in)

		return errors.Wrap(retryErr, "unable to get scangroup from client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return 0, err
	}
	return int(resp.GetOrgID()), nil
}
