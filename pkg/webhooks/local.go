package webhooks

import (
	"context"
	"time"
)

type LocalSender struct {
	c *Client
}

func NewLocalSender() *LocalSender {
	return &LocalSender{c: NewClient()}
}

func (s *LocalSender) Init() error {
	return nil
}

func (s *LocalSender) Send(ctx context.Context, events *Data) (*DataResponse, error) {
	code, err := s.c.SendEvent(ctx, events)
	if err != nil {
		return &DataResponse{Error: err.Error()}, err
	}

	return &DataResponse{StatusCode: code, DeliveredTime: time.Now().UnixNano()}, nil
}
