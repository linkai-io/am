package protoc

import (
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/dispatcher"

	context "golang.org/x/net/context"
)

type DispatcherProtocService struct {
	ds       am.DispatcherService
	reporter *load.RateReporter
}

func New(implementation am.DispatcherService, reporter *load.RateReporter) *DispatcherProtocService {
	return &DispatcherProtocService{ds: implementation, reporter: reporter}
}

func (s *DispatcherProtocService) PushAddresses(ctx context.Context, in *dispatcher.PushRequest) (*dispatcher.PushedResponse, error) {
	s.reporter.Increment(1)
	err := s.ds.PushAddresses(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	s.reporter.Increment(-1)
	return &dispatcher.PushedResponse{}, err
}
