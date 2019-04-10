package protoc

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/metrics/load"
	"github.com/linkai-io/am/protocservices/coordinator"

	context "golang.org/x/net/context"
)

type CoordProtocService struct {
	cs       am.CoordinatorService
	reporter *load.RateReporter
}

func New(implementation am.CoordinatorService, reporter *load.RateReporter) *CoordProtocService {
	return &CoordProtocService{cs: implementation, reporter: reporter}
}

func (s *CoordProtocService) StartGroup(ctx context.Context, in *coordinator.StartGroupRequest) (*coordinator.GroupStartedResponse, error) {
	s.reporter.Increment(1)
	err := s.cs.StartGroup(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	s.reporter.Increment(-1)
	return &coordinator.GroupStartedResponse{}, err
}

func (s *CoordProtocService) StopGroup(ctx context.Context, in *coordinator.StopGroupRequest) (*coordinator.GroupStoppedResponse, error) {
	s.reporter.Increment(1)
	response, err := s.cs.StopGroup(ctx, convert.UserContextToDomain(in.UserContext), int(in.OrgID), int(in.GroupID))
	s.reporter.Increment(-1)
	return &coordinator.GroupStoppedResponse{Message: response}, err
}
