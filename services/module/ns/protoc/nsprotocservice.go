package protoc

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/module"

	context "golang.org/x/net/context"
)

type NSProtocService struct {
	nsservice am.ModuleService
}

func New(implementation am.ModuleService) *NSProtocService {
	return &NSProtocService{nsservice: implementation}
}

func (s *NSProtocService) Analyze(ctx context.Context, in *module.AnalyzeRequest) (*module.AnalyzedResponse, error) {
	addresses, err := s.nsservice.Analyze(ctx, convert.AddressToDomain(in.Address))
	return &module.AnalyzedResponse{Addresses: addresses}, err
}
