package webhooks

import (
	"context"

	"github.com/linkai-io/am/am"
)

type Data struct {
	Settings *am.WebhookEventSettings `json:"settings"`
	Event    []*am.Event              `json:"events"`
}

type DataResponse struct {
	StatusCode    int    `json:"status_code"`
	DeliveredTime int64  `json:"delivery_time"`
	Data          *Data  `json:"data"`
	Error         string `json:"error"`
}

type Webhooker interface {
	Init() error
	Send(ctx context.Context, events *Data) (*DataResponse, error)
}

func New(env, region string) Webhooker {
	if env == "local" {
		return NewLocalSender()
	}
	return NewAWSSender(env, region)
}
