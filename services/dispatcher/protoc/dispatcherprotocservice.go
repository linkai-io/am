package protoc

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/dispatcher"

	context "golang.org/x/net/context"
)

type DispatcherProtocService struct {
	ds am.DispatcherService
}

func New(implementation am.DispatcherService) *DispatcherProtocService {
	return &DispatcherProtocService{ds: implementation}
}

func (d *DispatcherProtocService) PushAddresses(ctx context.Context, in *dispatcher.PushRequest) (*dispatcher.PushedResponse, error) {
	err := d.ds.PushAddresses(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	return &dispatcher.PushedResponse{}, err
}
