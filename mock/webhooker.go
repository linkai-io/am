package mock

import (
	"context"

	"github.com/linkai-io/am/pkg/webhooks"
)

type Webhooker struct {
	SendFn      func(ctx context.Context, events *webhooks.Data) (*webhooks.DataResponse, error)
	SendInvoked bool

	InitFn      func() error
	InitInvoked bool
}

func (w *Webhooker) Send(ctx context.Context, events *webhooks.Data) (*webhooks.DataResponse, error) {
	w.SendInvoked = true
	return w.SendFn(ctx, events)
}

func (w *Webhooker) Init() error {
	w.InitInvoked = true
	return w.InitFn()
}
