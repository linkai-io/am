package event

import (
	"context"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	service "github.com/linkai-io/am/protocservices/event"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

type Client struct {
	client         service.EventClient
	conn           *grpc.ClientConn
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 120)}
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.defaultTimeout = timeout
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.DialContext(context.Background(), "srv://consul/"+am.EventServiceKey, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = service.NewEventClient(conn)
	return nil
}

func (c *Client) Get(ctx context.Context, userContext am.UserContext, filter *am.EventFilter) ([]*am.Event, error) {
	var err error
	var resp *service.GetResponse

	in := &service.GetRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Filter:      convert.DomainToEventFilter(filter),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Get(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to get events from event client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return nil, err
	}

	return convert.EventsToDomain(resp.Events), nil
}

// GetSettings user settings
func (c *Client) GetSettings(ctx context.Context, userContext am.UserContext) (*am.UserEventSettings, error) {
	var err error
	var resp *service.GetSettingsResponse

	in := &service.GetSettingsRequest{
		UserContext: convert.DomainToUserContext(userContext),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.GetSettings(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to get user events settings from event client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return nil, err
	}

	return convert.UserEventSettingsToDomain(resp.Settings), nil
}

// MarkRead events
func (c *Client) MarkRead(ctx context.Context, userContext am.UserContext, notificationIDs []int64) error {
	var err error

	in := &service.MarkReadRequest{
		UserContext:     convert.DomainToUserContext(userContext),
		NotificationIDs: notificationIDs,
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		_, retryErr = c.client.MarkRead(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to mark events read from event client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return err
	}

	return nil
}

// Add events (system only?)
func (c *Client) Add(ctx context.Context, userContext am.UserContext, events []*am.Event) error {
	var err error

	in := &service.AddRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Data:        convert.DomainToUserEvents(events),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		_, retryErr = c.client.Add(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to add events from event client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return err
	}

	return nil
}

// UpdateSettings for user
func (c *Client) UpdateSettings(ctx context.Context, userContext am.UserContext, settings *am.UserEventSettings) error {
	var err error

	in := &service.UpdateSettingsRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Settings:    convert.DomainToUserEventSettings(settings),
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		_, retryErr = c.client.UpdateSettings(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to set settings for events from event client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return err
	}

	return nil
}

// NotifyComplete that a scan group has completed
func (c *Client) NotifyComplete(ctx context.Context, userContext am.UserContext, startTime int64, groupID int) error {
	var err error

	in := &service.NotifyCompleteRequest{
		UserContext: convert.DomainToUserContext(userContext),
		GroupID:     int32(groupID),
		StartTime:   startTime,
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		_, retryErr = c.client.NotifyComplete(ctxDeadline, in)
		return errors.Wrap(retryErr, "unable to notify complete from event client")
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return err
	}

	return nil
}
