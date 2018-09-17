package protoc

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/coordinator"

	context "golang.org/x/net/context"
)

type CoordProtocService struct {
	cs am.CoordinatorService
}

func New(implementation am.CoordinatorService) *CoordProtocService {
	return &CoordProtocService{cs: implementation}
}

func (o *CoordProtocService) StartGroup(ctx context.Context, in *coordinator.StartGroupRequest) (*coordinator.GroupStartedResponse, error) {
	err := o.cs.StartGroup(ctx, convert.UserContextToDomain(in.UserContext), int(in.GroupID))
	return &coordinator.GroupStartedResponse{}, err
}

func (o *CoordProtocService) Register(ctx context.Context, in *coordinator.RegisterRequest) (*coordinator.RegisteredResponse, error) {
	err := o.cs.Register(ctx, in.DispatcherAddr, in.DispatcherID)
	return &coordinator.RegisteredResponse{}, err
}
